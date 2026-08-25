package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/vialfill/internal/model"
)

func (a *App) WarmupStatus() (ready bool, detail string) {
	snap := a.Snapshot()
	if snap.Dosepump.FlushStartedAt.IsZero() {
		return false, "flush not started"
	}
	if !a.flushWindow.Ready(snap.Dosepump.FlushStartedAt) {
		return false, "flush window open"
	}
	if !snap.Dosepump.IgnitionAt.IsZero() && !a.warmupWindow.Ready(snap.Dosepump.IgnitionAt) {
		return false, "dosepump warmup window open"
	}
	if !snap.Vialtray.LastSwellAt.IsZero() {
		if err := a.vialtray.RequireSettled(snap.Vialtray); err != nil {
			return false, "vialtray swell settling"
		}
	}
	return true, "ready"
}

func (a *App) WaitWarmup(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w", model.ErrContextDone)
		default:
		}
		ready, _ := a.WarmupStatus()
		if ready {
			return nil
		}
	}
}

func (a *App) FlushRemaining() string {
	snap := a.Snapshot()
	if snap.Dosepump.FlushStartedAt.IsZero() {
		return "not started"
	}
	if a.flushWindow.Ready(snap.Dosepump.FlushStartedAt) {
		return "complete"
	}
	return "in progress"
}

func (a *App) DosepumpWarmupRemaining() string {
	snap := a.Snapshot()
	if snap.Dosepump.IgnitionAt.IsZero() {
		return "not ignited"
	}
	if a.warmupWindow.Ready(snap.Dosepump.IgnitionAt) {
		return "complete"
	}
	return "in progress"
}
