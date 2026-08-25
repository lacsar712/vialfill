package dosepump

import (
	"math"

	"github.com/lacsar712/vialfill/internal/clock"
	"github.com/lacsar712/vialfill/internal/model"
)

type BurnerController struct {
	clk clock.ProcessClock
}

func NewBurnerController(clk clock.ProcessClock) *BurnerController {
	return &BurnerController{clk: clk}
}

func (b *BurnerController) EstimateIsolatorTemp(reading model.DosepumpReading) float64 {
	base := 300.0
	sterileHeat := reading.SterileFlowTPH * 50
	airCool := reading.AirflowTPH * 2
	return base + sterileHeat - airCool
}

func (b *BurnerController) FillStable(reading model.DosepumpReading) bool {
	if reading.BurnerPhase != model.BurnerStable && reading.BurnerPhase != model.BurnerIgnition {
		return false
	}
	return reading.IsolatorTempF > 800 && reading.ExcessO2Pct >= model.MinIsolatorO2Percent
}

func (b *BurnerController) TripRequired(reading model.DosepumpReading) bool {
	if reading.ExcessO2Pct > model.MaxIsolatorO2Percent*2 {
		return true
	}
	if reading.BurnerPhase == model.BurnerTrip {
		return true
	}
	if reading.IsolatorTempF > 3500 {
		return true
	}
	return false
}

func (b *BurnerController) PhaseLabel(phase model.BurnerPhase) string {
	switch phase {
	case model.BurnerIdle:
		return "Idle"
	case model.BurnerFlush:
		return "Flush"
	case model.BurnerIgnition:
		return "Ignition"
	case model.BurnerStable:
		return "Stable Fill"
	case model.BurnerTrip:
		return "Tripped"
	default:
		return string(phase)
	}
}

func (b *BurnerController) HeatReleaseMW(reading model.DosepumpReading) float64 {
	return reading.SterileFlowTPH * 12.5
}

func (b *BurnerController) TurndownRatio(settings model.PlantSettings, currentSterile float64) float64 {
	if settings.SterileFlowTPH <= 0 {
		return 0
	}
	return currentSterile / settings.SterileFlowTPH
}

func (b *BurnerController) MinStableSterile(settings model.PlantSettings) float64 {
	return settings.SterileFlowTPH * 0.25
}

func (b *BurnerController) NormalizeSterile(flow, max float64) float64 {
	return math.Min(math.Max(flow, 0), max)
}
