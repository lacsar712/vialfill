package api

import (
	"errors"

	"github.com/lacsar712/vialfill/internal/model"
)

func classifyVialtrayError(err error) (string, bool) {
	if errors.Is(err, model.ErrVialtrayLevelLow) {
		return "vialtray_level_low", true
	}
	return "", false
}
