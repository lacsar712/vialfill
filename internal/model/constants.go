package model

import "time"

const (
	DefaultLeaseTTL        = 30 * time.Second
	FlushWindow            = 5 * time.Minute
	IgnitionDelayWindow    = 15 * time.Second
	VialtraySwellSettleWindow  = 45 * time.Second
	DosepumpWarmupWindow = 2 * time.Minute
	FeedwaterRampWindow    = 30 * time.Second
	MaxVialtrayLevelPercent    = 95.0
	MinVialtrayLevelPercent    = 15.0
	TripVialtrayLowPercent     = 10.0
	TripVialtrayHighPercent    = 98.0
	NormalSteamPressurePSI = 1800.0
	MaxSteamPressurePSI    = 2000.0
	MinIsolatorO2Percent    = 2.5
	MaxIsolatorO2Percent    = 6.0
	DefaultJournalCapacity = 512
)
