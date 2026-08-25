package fsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/vialfill/internal/model"
)

type FilllaneFSM struct {
	mu            sync.RWMutex
	state         model.PlantState
	sterilePermissive bool
	flushComplete  bool
	hooks          *HookChain
}

func NewFilllaneFSM(unitID string) *FilllaneFSM {
	_ = unitID
	return &FilllaneFSM{state: model.StateColdStandby, hooks: NewHookChain()}
}

func (f *FilllaneFSM) Hooks() *HookChain { return f.hooks }

func (f *FilllaneFSM) State() model.PlantState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *FilllaneFSM) SetSterilePermissive(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sterilePermissive = ok
}

func (f *FilllaneFSM) SetFlushComplete(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flushComplete = ok
}

func (f *FilllaneFSM) SterilePermissive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.sterilePermissive
}

func (f *FilllaneFSM) Dispatch(ctx context.Context, event PlantEvent) (model.PlantState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-ctx.Done():
		return f.state, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if event == EvTrip {
		from := f.state
		if f.hooks != nil {
			if err := f.hooks.RunBefore(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		f.state = model.StateTrip
		if f.hooks != nil {
			if err := f.hooks.RunAfter(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		return f.state, nil
	}
	next, ok := NextState(f.state, event)
	if !ok {
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalTransition)
	}
	if event == EvIgnite && !f.sterilePermissive {
		return f.state, fmt.Errorf("%w", model.ErrSterilePermissive)
	}
	if event == EvFlushComplete && !f.flushComplete {
		return f.state, fmt.Errorf("%w", model.ErrFlushIncomplete)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func (f *FilllaneFSM) ForceState(state model.PlantState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
