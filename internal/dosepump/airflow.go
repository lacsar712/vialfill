package dosepump

import (
	"math"

	"github.com/lacsar712/vialfill/internal/clock"
	"github.com/lacsar712/vialfill/internal/model"
)

type AirflowBalancer struct {
	clk clock.ProcessClock
}

func NewAirflowBalancer(clk clock.ProcessClock) *AirflowBalancer {
	return &AirflowBalancer{clk: clk}
}

func (a *AirflowBalancer) FlushRate() float64 { return 120.0 }

func (a *AirflowBalancer) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.SterileFlowTPH * 18
}

func (a *AirflowBalancer) Compute(snap model.PlantSnapshot) float64 {
	stoich := snap.Dosepump.SterileFlowTPH * 15.5
	o2Factor := 1 + (snap.Settings.ExcessO2Setpoint / 100)
	return stoich * o2Factor
}

func (a *AirflowBalancer) ExcessO2(reading model.DosepumpReading) float64 {
	if reading.AirflowTPH <= 0 {
		return 0
	}
	stoichAir := reading.SterileFlowTPH * 15.5
	if stoichAir <= 0 {
		return 21.0
	}
	excess := (reading.AirflowTPH - stoichAir) / stoichAir
	return math.Max(0, excess*100/5)
}

func (a *AirflowBalancer) WithinLimits(reading model.DosepumpReading) bool {
	o2 := reading.ExcessO2Pct
	return o2 >= model.MinIsolatorO2Percent && o2 <= model.MaxIsolatorO2Percent
}

func (a *AirflowBalancer) TrimAirflow(current, targetO2, actualO2 float64) float64 {
	err := targetO2 - actualO2
	return current * (1 + err*0.02)
}

func (a *AirflowBalancer) DamperPosition(airflowTPH, maxTPH float64) float64 {
	if maxTPH <= 0 {
		return 0
	}
	return math.Min(100, (airflowTPH/maxTPH)*100)
}

func (a *AirflowBalancer) FDFanPowerKW(airflowTPH float64) float64 {
	return airflowTPH * 0.8
}
