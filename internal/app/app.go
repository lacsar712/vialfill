package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/vialfill/internal/filllane"
	"github.com/lacsar712/vialfill/internal/clock"
	"github.com/lacsar712/vialfill/internal/dosepump"
	"github.com/lacsar712/vialfill/internal/config"
	"github.com/lacsar712/vialfill/internal/vialtray"
	"github.com/lacsar712/vialfill/internal/fsm"
	"github.com/lacsar712/vialfill/internal/interlock"
	"github.com/lacsar712/vialfill/internal/model"
	"github.com/lacsar712/vialfill/internal/store"
)

type App struct {
	cfg           config.Config
	clk           clock.ProcessClock
	store         *store.PlantStore
	journal       *store.Journal
	fsm           *fsm.FilllaneFSM
	filllane        *filllane.Controller
	dosepump    *dosepump.Coordinator
	vialtray          *vialtray.Coordinator
	interlock     *interlock.Interlock
	permissives   *interlock.PermissiveSet
	coordLock     *interlock.CoordinationLock
	scheduler     *clock.Scheduler
	flushWindow   *clock.FlushWindow
	warmupWindow  *clock.DosepumpWarmupWindow
	telemetry     *Telemetry
	tickCancels    map[string]context.CancelFunc
	sterileLoopCancels map[string]context.CancelFunc
	mu             sync.RWMutex
}

func New(cfg config.Config, clk clock.ProcessClock) *App {
	return &App{
		cfg:          cfg,
		clk:          clk,
		store:        store.NewPlantStore(),
		journal:      store.NewJournal(cfg.JournalPath, cfg.JournalCapacity),
		fsm:          fsm.NewFilllaneFSM(cfg.UnitID),
		filllane:       filllane.NewController(clk),
		dosepump:   dosepump.NewCoordinator(clk),
		vialtray:         vialtray.NewCoordinator(clk),
		interlock:    interlock.NewInterlock(cfg.LeaseTTL),
		permissives:  interlock.NewPermissiveSet(),
		coordLock:    interlock.NewCoordinationLock(),
		scheduler:    clock.NewScheduler(clk),
		flushWindow:  clock.NewFlushWindow(clk),
		warmupWindow: clock.NewDosepumpWarmupWindow(clk),
		telemetry:    NewTelemetry(cfg.UnitID),
		tickCancels:     make(map[string]context.CancelFunc),
		sterileLoopCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) Snapshot() model.PlantSnapshot {
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return model.DefaultSnapshot(a.cfg.UnitID)
	}
	return snap
}

func (a *App) Config() config.Config              { return a.cfg }
func (a *App) Clock() clock.ProcessClock          { return a.clk }
func (a *App) FSM() *fsm.FilllaneFSM                { return a.fsm }
func (a *App) UnitID() string                     { return a.cfg.UnitID }
func (a *App) Store() *store.PlantStore           { return a.store }
func (a *App) Interlock() *interlock.Interlock    { return a.interlock }
func (a *App) Telemetry() TelemetrySnapshot       { return a.telemetry.Snapshot() }
func (a *App) Journal() *store.Journal            { return a.journal }

func (a *App) journalEvent(ev, payload string) {
	_, _ = a.journal.Append(a.cfg.UnitID, ev, payload)
}

func (a *App) syncState(state model.PlantState) {
	_ = a.store.UpdateState(a.cfg.UnitID, state)
}

func (a *App) isFiring(state model.PlantState) bool {
	return state == model.StateFiring || state == model.StateLoadFollow || state == model.StateRamp
}

func (a *App) refreshPermissives(snap model.PlantSnapshot) {
	a.permissives.SetVialtray(a.vialtray.Level().WithinLimits(snap.Vialtray.LevelPercent))
	a.permissives.SetPressure(a.filllane.Pressure().WithinTripLimits(snap.Filllane.SteamPressurePSI, a.isFiring(snap.State)))
	a.permissives.SetDosepump(a.dosepump.Burner().FillStable(snap.Dosepump))
	a.permissives.SetSterile(snap.Dosepump.SterileFlowTPH > 0 || snap.State == model.StateFlush)
	a.permissives.SetIgnition(snap.Dosepump.BurnerPhase == model.BurnerStable || snap.Dosepump.BurnerPhase == model.BurnerIgnition)
	a.fsm.SetSterilePermissive(a.permissives.SterileOK())
	a.fsm.SetFlushComplete(a.flushWindow.Ready(snap.Dosepump.FlushStartedAt))
}

func (a *App) tickLabel() string {
	return fmt.Sprintf("%s-tick", a.cfg.UnitID)
}
