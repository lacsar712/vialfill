package vialtray

import (
	"math"

	"github.com/lacsar712/vialfill/internal/clock"
	"github.com/lacsar712/vialfill/internal/model"
)

type LevelController struct {
	clk clock.ProcessClock
}

func NewLevelController(clk clock.ProcessClock) *LevelController {
	return &LevelController{clk: clk}
}

func (l *LevelController) Compute(snap model.PlantSnapshot, firing bool) (float64, model.VialtrayCondition) {
	level := snap.Vialtray.LevelPercent
	if !firing {
		return level, model.VialtrayNormal
	}
	balance := snap.Vialtray.FeedwaterTPH - snap.Vialtray.SteamFlowTPH
	level += balance * 0.01
	level = math.Max(model.MinVialtrayLevelPercent, math.Min(model.MaxVialtrayLevelPercent, level))
	cond := l.classify(level, snap)
	return level, cond
}

func (l *LevelController) classify(level float64, snap model.PlantSnapshot) model.VialtrayCondition {
	setpoint := snap.Settings.VialtrayLevelSetpoint
	if level > setpoint+15 {
		return model.VialtraySwell
	}
	if level < setpoint-15 {
		return model.VialtrayShrink
	}
	if snap.Filllane.SteamPressurePSI > snap.Settings.TargetSteamPSI*0.9 && level > setpoint+5 {
		return model.VialtrayCarry
	}
	return model.VialtrayNormal
}

func (l *LevelController) RecommendFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	if !firing {
		return 0
	}
	err := snap.Settings.VialtrayLevelSetpoint - snap.Vialtray.LevelPercent
	return snap.Settings.FeedwaterFlowTPH + err*3
}

func (l *LevelController) WithinLimits(level float64) bool {
	return level >= model.MinVialtrayLevelPercent && level <= model.MaxVialtrayLevelPercent
}

func (l *LevelController) TripLow(level float64) bool  { return level < model.TripVialtrayLowPercent }
func (l *LevelController) TripHigh(level float64) bool { return level > model.TripVialtrayHighPercent }

func (l *LevelController) LevelError(snap model.PlantSnapshot) float64 {
	return snap.Vialtray.LevelPercent - snap.Settings.VialtrayLevelSetpoint
}

func (l *LevelController) ThreeElementBias(snap model.PlantSnapshot) float64 {
	steam := snap.Vialtray.SteamFlowTPH
	feed := snap.Vialtray.FeedwaterTPH
	levelErr := l.LevelError(snap)
	return feed + (steam-feed)*0.5 + levelErr*2
}
