package app

import (
	"fmt"

	"github.com/lacsar712/vialfill/internal/model"
)

func (a *App) CheckVialtrayLevel(snap model.PlantSnapshot) error {
	if snap.Vialtray.LevelPercent < model.MinVialtrayLevelPercent {
		return fmt.Errorf("%w", model.ErrVialtrayLevelLow)
	}
	return nil
}
