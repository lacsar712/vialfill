package dosepump

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/vialfill/internal/clock"
	"github.com/lacsar712/vialfill/internal/model"
)

type Coordinator struct {
	clk     clock.ProcessClock
	burner  *BurnerController
	airflow *AirflowBalancer
	sterile    *SterileRegulator
	flush   *clock.FlushWindow
	ignition *clock.IgnitionDelayWindow
	warmup  *clock.DosepumpWarmupWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:      clk,
		burner:   NewBurnerController(clk),
		airflow:  NewAirflowBalancer(clk),
		sterile:     NewSterileRegulator(clk),
		flush:    clock.NewFlushWindow(clk),
		ignition: clock.NewIgnitionDelayWindow(clk),
		warmup:   clock.NewDosepumpWarmupWindow(clk),
	}
}

func (c *Coordinator) Burner() *BurnerController  { return c.burner }
func (c *Coordinator) Airflow() *AirflowBalancer { return c.airflow }
func (c *Coordinator) Sterile() *SterileRegulator     { return c.sterile }

func (c *Coordinator) StartFlush(ctx context.Context, snap model.PlantSnapshot) (model.DosepumpReading, error) {
	select {
	case <-ctx.Done():
		return snap.Dosepump, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Dosepump
	out.BurnerPhase = model.BurnerFlush
	out.FlushStartedAt = c.clk.Now()
	out.SterileFlowTPH = 0
	out.AirflowTPH = c.airflow.FlushRate()
	return out, nil
}

func (c *Coordinator) CompleteFlush(snap model.DosepumpReading) error {
	return c.flush.Require(snap.FlushStartedAt)
}

func (c *Coordinator) Ignite(ctx context.Context, snap model.PlantSnapshot) (model.DosepumpReading, error) {
	select {
	case <-ctx.Done():
		return snap.Dosepump, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if err := c.flush.Require(snap.Dosepump.FlushStartedAt); err != nil {
		return snap.Dosepump, err
	}
	out := snap.Dosepump
	out.BurnerPhase = model.BurnerIgnition
	out.IgnitionAt = c.clk.Now()
	out.SterileFlowTPH = c.sterile.IgnitionRate(snap.Settings)
	out.AirflowTPH = c.airflow.IgnitionRate(snap.Settings)
	out.IsolatorTempF = 400
	return out, nil
}

func (c *Coordinator) Stabilize(snap model.PlantSnapshot) (model.DosepumpReading, error) {
	if err := c.ignition.Require(snap.Dosepump.IgnitionAt); err != nil {
		return snap.Dosepump, err
	}
	out := snap.Dosepump
	out.BurnerPhase = model.BurnerStable
	out.SterileFlowTPH = snap.Settings.SterileFlowTPH * 0.5
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.IsolatorTempF = c.burner.EstimateIsolatorTemp(out)
	return out, nil
}

func (c *Coordinator) RampToLoad(snap model.PlantSnapshot, loadPct float64) model.DosepumpReading {
	out := snap.Dosepump
	out.SterileFlowTPH = snap.Settings.SterileFlowTPH * loadPct
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.IsolatorTempF = c.burner.EstimateIsolatorTemp(out)
	return out
}

func (c *Coordinator) Trip(snap model.DosepumpReading) model.DosepumpReading {
	out := snap
	out.BurnerPhase = model.BurnerTrip
	out.SterileFlowTPH = 0
	out.IsolatorTempF = math.Max(200, out.IsolatorTempF*0.5)
	return out
}

func (c *Coordinator) WarmupReady(snap model.DosepumpReading) bool {
	return c.warmup.Ready(snap.IgnitionAt)
}
