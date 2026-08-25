package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/vialfill/internal/model"
)

const maxCipdrainOpeningPct = 100.0

func (a *App) OpenCipdrain(ctx context.Context, holder string, openingPct float64) error {
	_ = holder
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if openingPct >= maxCipdrainOpeningPct {
		return fmt.Errorf("cipdrain: %w", model.ErrCipdrainLimit)
	}
	return nil
}

func (a *App) CipdrainAfterShutdown(ctx context.Context, openingPct float64) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	snap := a.Snapshot()
	if snap.State != model.StateTrip && snap.State != model.StateColdStandby {
		return fmt.Errorf("plant not shut down")
	}
	if openingPct >= maxCipdrainOpeningPct {
		return fmt.Errorf("unknown fault")
	}
	return nil
}
