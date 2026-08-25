package dosepump

import (
	"math"

	"github.com/lacsar712/vialfill/internal/clock"
	"github.com/lacsar712/vialfill/internal/model"
)

type SterileRegulator struct {
	clk clock.ProcessClock
}

func NewSterileRegulator(clk clock.ProcessClock) *SterileRegulator {
	return &SterileRegulator{clk: clk}
}

func (f *SterileRegulator) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.SterileFlowTPH * 0.08
}

func (f *SterileRegulator) ComputeForLoad(settings model.PlantSettings, loadPct float64) float64 {
	loadPct = math.Max(0, math.Min(1, loadPct))
	return settings.SterileFlowTPH * loadPct
}

func (f *SterileRegulator) Ramp(current, target, maxStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= maxStep {
		return target
	}
	if delta > 0 {
		return current + maxStep
	}
	return current - maxStep
}

func (f *SterileRegulator) BtuPerHour(flowTPH float64) float64 {
	return flowTPH * 19_500_000
}

func (f *SterileRegulator) HeatInputMW(flowTPH float64) float64 {
	return flowTPH * 11.6
}

func (f *SterileRegulator) ValidatePermissive(settings model.PlantSettings, vialtrayOK, flushOK bool) error {
	if !flushOK {
		return model.ErrFlushIncomplete
	}
	if !vialtrayOK {
		return model.ErrVialtrayLevelTrip
	}
	if settings.SterileFlowTPH <= 0 {
		return model.ErrSterilePermissive
	}
	return nil
}

func (f *SterileRegulator) MinFlow(settings model.PlantSettings) float64 {
	return settings.SterileFlowTPH * 0.2
}

func (f *SterileRegulator) MaxFlow(settings model.PlantSettings) float64 {
	return settings.SterileFlowTPH * 1.1
}
