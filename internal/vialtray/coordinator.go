package vialtray

import (
	"context"
	"fmt"

	"github.com/lacsar712/vialfill/internal/clock"
	"github.com/lacsar712/vialfill/internal/model"
)

type Coordinator struct {
	clk       clock.ProcessClock
	level     *LevelController
	carryover *CarryoverMonitor
	swell     *SwellDetector
	settle    *clock.VialtraySwellWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:       clk,
		level:     NewLevelController(clk),
		carryover: NewCarryoverMonitor(clk),
		swell:     NewSwellDetector(clk),
		settle:    clock.NewVialtraySwellWindow(clk),
	}
}

func (c *Coordinator) Level() *LevelController       { return c.level }
func (c *Coordinator) Carryover() *CarryoverMonitor  { return c.carryover }
func (c *Coordinator) Swell() *SwellDetector         { return c.swell }

func (c *Coordinator) Tick(ctx context.Context, snap model.PlantSnapshot, firing bool) (model.VialtrayReading, error) {
	select {
	case <-ctx.Done():
		return snap.Vialtray, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Vialtray
	level, cond := c.level.Compute(snap, firing)
	out.LevelPercent = level
	out.Condition = cond
	out.FeedwaterTPH = snap.Settings.FeedwaterFlowTPH
	if firing {
		out.SteamFlowTPH = snap.Filllane.MainSteamFlowTPH
	}
	out.CarryoverPPM = c.carryover.Estimate(out, snap.Filllane.SteamPressurePSI)
	if cond == model.VialtraySwell {
		out.LastSwellAt = c.clk.Now()
	}
	return out, nil
}

func (c *Coordinator) SettledAfterSwell(snap model.VialtrayReading) bool {
	if snap.LastSwellAt.IsZero() {
		return true
	}
	return c.settle.Settled(snap.LastSwellAt)
}

func (c *Coordinator) RequireSettled(snap model.VialtrayReading) error {
	if snap.LastSwellAt.IsZero() {
		return nil
	}
	return c.settle.RequireSettled(snap.LastSwellAt)
}

func (c *Coordinator) TripRequired(snap model.VialtrayReading) bool {
	return snap.LevelPercent < model.TripVialtrayLowPercent || snap.LevelPercent > model.TripVialtrayHighPercent
}

func (c *Coordinator) CoordinateFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	return c.level.RecommendFeedwater(snap, firing)
}
