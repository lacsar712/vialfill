package clock

import (
	"time"

	"github.com/lacsar712/vialfill/internal/model"
)

type Window struct{ Duration time.Duration }

func NewWindow(d time.Duration) Window {
	if d < 0 {
		d = 0
	}
	return Window{Duration: d}
}

func (w Window) Satisfied(clk ProcessClock, anchor time.Time) bool {
	if w.Duration == 0 {
		return true
	}
	return clk.Since(anchor) >= w.Duration
}

func (w Window) WaitUntil(clk ProcessClock, anchor time.Time) error {
	if w.Satisfied(clk, anchor) {
		return nil
	}
	return model.ErrWindowOpen
}

type FlushWindow struct {
	clk    ProcessClock
	window Window
}

func NewFlushWindow(clk ProcessClock) *FlushWindow {
	return &FlushWindow{clk: clk, window: NewWindow(model.FlushWindow)}
}

func (p *FlushWindow) Ready(startedAt time.Time) bool {
	return p.window.Satisfied(p.clk, startedAt)
}

func (p *FlushWindow) Require(startedAt time.Time) error {
	if p.Ready(startedAt) {
		return nil
	}
	return model.ErrFlushIncomplete
}

type IgnitionDelayWindow struct {
	clk    ProcessClock
	window Window
}

func NewIgnitionDelayWindow(clk ProcessClock) *IgnitionDelayWindow {
	return &IgnitionDelayWindow{clk: clk, window: NewWindow(model.IgnitionDelayWindow)}
}

func (i *IgnitionDelayWindow) Ready(ignitionAt time.Time) bool {
	return i.window.Satisfied(i.clk, ignitionAt)
}

func (i *IgnitionDelayWindow) Require(ignitionAt time.Time) error {
	if i.Ready(ignitionAt) {
		return nil
	}
	return model.ErrWindowOpen
}

type VialtraySwellWindow struct {
	clk    ProcessClock
	window Window
}

func NewVialtraySwellWindow(clk ProcessClock) *VialtraySwellWindow {
	return &VialtraySwellWindow{clk: clk, window: NewWindow(model.VialtraySwellSettleWindow)}
}

func (d *VialtraySwellWindow) Settled(swellAt time.Time) bool {
	return d.window.Satisfied(d.clk, swellAt)
}

func (d *VialtraySwellWindow) RequireSettled(swellAt time.Time) error {
	if d.Settled(swellAt) {
		return nil
	}
	return model.ErrWindowOpen
}

type DosepumpWarmupWindow struct {
	clk    ProcessClock
	window Window
}

func NewDosepumpWarmupWindow(clk ProcessClock) *DosepumpWarmupWindow {
	return &DosepumpWarmupWindow{clk: clk, window: NewWindow(model.DosepumpWarmupWindow)}
}

func (c *DosepumpWarmupWindow) Ready(ignitionAt time.Time) bool {
	return c.window.Satisfied(c.clk, ignitionAt)
}

func (c *DosepumpWarmupWindow) Require(ignitionAt time.Time) error {
	if c.Ready(ignitionAt) {
		return nil
	}
	return model.ErrWindowOpen
}

type FeedwaterRampTracker struct {
	clk    ProcessClock
	window Window
	anchor time.Time
}

func NewFeedwaterRampTracker(clk ProcessClock) *FeedwaterRampTracker {
	return &FeedwaterRampTracker{
		clk:    clk,
		window: NewWindow(model.FeedwaterRampWindow),
		anchor: clk.Now(),
	}
}

func (f *FeedwaterRampTracker) Reset()          { f.anchor = f.clk.Now() }
func (f *FeedwaterRampTracker) Satisfied() bool { return f.window.Satisfied(f.clk, f.anchor) }
func (f *FeedwaterRampTracker) Require() error  { return f.window.WaitUntil(f.clk, f.anchor) }
