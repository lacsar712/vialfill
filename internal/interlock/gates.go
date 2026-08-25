package interlock

import (
	"fmt"

	"github.com/lacsar712/vialfill/internal/model"
)

type PermissiveSet struct {
	sterileOK       bool
	ignitionOK   bool
	vialtrayOK       bool
	pressureOK   bool
	dosepumpOK bool
}

func NewPermissiveSet() *PermissiveSet { return &PermissiveSet{} }

func (p *PermissiveSet) SetSterile(ok bool)       { p.sterileOK = ok }
func (p *PermissiveSet) SetIgnition(ok bool)   { p.ignitionOK = ok }
func (p *PermissiveSet) SetVialtray(ok bool)       { p.vialtrayOK = ok }
func (p *PermissiveSet) SetPressure(ok bool)   { p.pressureOK = ok }
func (p *PermissiveSet) SetDosepump(ok bool) { p.dosepumpOK = ok }

func (p *PermissiveSet) SterileOK() bool       { return p.sterileOK }
func (p *PermissiveSet) IgnitionOK() bool   { return p.ignitionOK }
func (p *PermissiveSet) VialtrayOK() bool       { return p.vialtrayOK }
func (p *PermissiveSet) PressureOK() bool   { return p.pressureOK }
func (p *PermissiveSet) DosepumpOK() bool { return p.dosepumpOK }

func (p *PermissiveSet) AllFiring() bool {
	return p.sterileOK && p.ignitionOK && p.vialtrayOK && p.pressureOK && p.dosepumpOK
}

func (p *PermissiveSet) CheckIgnition() error {
	if !p.sterileOK {
		return fmt.Errorf("%w", model.ErrSterilePermissive)
	}
	if !p.ignitionOK {
		return fmt.Errorf("%w", model.ErrIgnitionBlocked)
	}
	return nil
}

func CheckFillLoss(reading model.DosepumpReading) error {
	if reading.BurnerPhase == model.BurnerStable && reading.IsolatorTempF < 600 {
		return fmt.Errorf("%w", model.ErrFillLoss)
	}
	return nil
}

func (p *PermissiveSet) CheckFiring() error {
	if err := p.CheckIgnition(); err != nil {
		return err
	}
	if !p.vialtrayOK {
		return fmt.Errorf("%w", model.ErrVialtrayLevelTrip)
	}
	if !p.pressureOK {
		return fmt.Errorf("%w", model.ErrPressureTrip)
	}
	if !p.dosepumpOK {
		return fmt.Errorf("%w", model.ErrDosepumpTrip)
	}
	return nil
}

type CoordinationLock struct {
	holder string
	held   bool
}

func NewCoordinationLock() *CoordinationLock { return &CoordinationLock{} }

func (c *CoordinationLock) Acquire(holder string) error {
	if c.held {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	c.holder = holder
	c.held = true
	return nil
}

func (c *CoordinationLock) Release(holder string) {
	if c.held && c.holder == holder {
		c.held = false
		c.holder = ""
	}
}

func (c *CoordinationLock) Require(holder string) error {
	if !c.held || c.holder != holder {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	return nil
}

func (c *CoordinationLock) Held() bool { return c.held }
