package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/vialfill/internal/clock"
	"github.com/lacsar712/vialfill/internal/model"
)

func (a *App) advanceClock(d time.Duration) {
	if mc, ok := a.clk.(*clock.ManualClock); ok {
		mc.Advance(d)
		time.Sleep(time.Millisecond)
	} else {
		time.Sleep(d)
	}
}

func (a *App) bindSterileLoop(holder string, ctx context.Context) context.Context {
	a.mu.Lock()
	if cancel, ok := a.sterileLoopCancels[holder]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.sterileLoopCancels[holder] = cancel
	a.mu.Unlock()
	return child
}

func (a *App) cancelSterileLoop(holder string) {
	a.mu.Lock()
	if cancel, ok := a.sterileLoopCancels[holder]; ok {
		cancel()
		delete(a.sterileLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) cancelAllSterileLoops() {
	a.mu.Lock()
	for holder, cancel := range a.sterileLoopCancels {
		cancel()
		delete(a.sterileLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) CoalFeedTPH() float64 {
	return a.Snapshot().Dosepump.SterileFlowTPH
}

func (a *App) RunSterileRamp(ctx context.Context, holder string, targetTPH float64) error {
	loopCtx := a.bindSterileLoop(holder, ctx)
	defer a.cancelSterileLoop(holder)
	for {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		current := snap.Dosepump.SterileFlowTPH
		if current >= targetTPH {
			return nil
		}
		comb := snap.Dosepump
		comb.SterileFlowTPH = current + 1.0
		_ = a.store.UpdateDosepump(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.SterileFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
}

func (a *App) RunCoalFeed(ctx context.Context, holder string, steps int) error {
	loopCtx := a.bindSterileLoop(holder, ctx)
	defer a.cancelSterileLoop(holder)
	for i := 0; steps <= 0 || i < steps; i++ {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		comb := snap.Dosepump
		comb.SterileFlowTPH += 0.5
		_ = a.store.UpdateDosepump(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.SterileFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}
