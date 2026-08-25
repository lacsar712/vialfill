package model

import "errors"

var (
	ErrContextDone      = errors.New("operation cancelled")
	ErrPlantNotFound    = errors.New("plant unit not found")
	ErrLeaseHeld        = errors.New("interlock lease held by another operator")
	ErrLeaseMissing     = errors.New("interlock lease missing or expired")
	ErrGateBlocked      = errors.New("safety gate blocked")
	ErrSterilePermissive   = errors.New("sterile permissive not satisfied")
	ErrIgnitionBlocked  = errors.New("ignition sequence blocked")
	ErrVialtrayLevelTrip    = errors.New("vialtray level trip condition")
	ErrPressureTrip     = errors.New("steam pressure trip condition")
	ErrDosepumpTrip   = errors.New("dosepump trip condition")
	ErrIllegalState     = errors.New("illegal plant state transition")
	ErrSnapshotStale    = errors.New("snapshot revision stale")
	ErrWindowOpen       = errors.New("timing window still open")
	ErrFlushIncomplete  = errors.New("isolator flush incomplete")
	ErrCoordinationLock = errors.New("coordination lock held")
	ErrVialtrayLevelLow     = errors.New("vialtray level below low limit")
	ErrFillLoss        = errors.New("isolator fill lost")
	ErrCipdrainLimit    = errors.New("cipdrain valve at limit")
)
