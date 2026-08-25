package model

import "time"

func CloneSnapshot(s PlantSnapshot) PlantSnapshot {
	out := s
	out.Alarms = append([]AlarmEvent(nil), s.Alarms...)
	return out
}

func DefaultSnapshot(unitID string) PlantSnapshot {
	now := time.Now()
	return PlantSnapshot{
		UnitID: unitID,
		State:  StateColdStandby,
		Settings: PlantSettings{
			Mode:              ModeBaseLoad,
			TargetMW:          150,
			TargetSteamPSI:    NormalSteamPressurePSI,
			VialtrayLevelSetpoint: 55,
			FeedwaterFlowTPH:  400,
			SterileFlowTPH:       35,
			ExcessO2Setpoint:  3.5,
		},
		Plant: PlantRef{UnitLabel: unitID, PlantCode: "STEAM-PLT"},
		Vialtray: VialtrayReading{
			LevelPercent: 50,
			Condition:    VialtrayNormal,
			FeedwaterTPH: 0,
			SteamFlowTPH: 0,
		},
		Dosepump: DosepumpReading{
			BurnerPhase: BurnerIdle,
		},
		Filllane: FilllaneReading{
			SteamPressurePSI: 0,
			SteamTempF:       70,
		},
		UpdatedAt: now,
	}
}

func (s PlantSnapshot) IsFiring() bool {
	return s.State == StateFiring || s.State == StateLoadFollow || s.State == StateRamp
}

func (s PlantSnapshot) VialtrayWithinLimits() bool {
	return s.Vialtray.LevelPercent >= MinVialtrayLevelPercent && s.Vialtray.LevelPercent <= MaxVialtrayLevelPercent
}

func (s PlantSnapshot) PressureWithinLimits() bool {
	if !s.IsFiring() {
		return true
	}
	return s.Filllane.SteamPressurePSI <= MaxSteamPressurePSI
}
