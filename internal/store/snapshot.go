package store

import "github.com/lacsar712/vialfill/internal/model"

type VialtraySnapshotView struct {
	UnitID   string
	Vialtray     model.VialtrayReading
	Alarms   []model.AlarmEvent
	Revision uint64
}

func CloneVialtraySnapshot(s model.PlantSnapshot) VialtraySnapshotView {
	out := VialtraySnapshotView{
		UnitID:   s.UnitID,
		Vialtray:     s.Vialtray,
		Revision: s.Revision,
	}
	out.Alarms = s.Alarms[:len(s.Alarms):len(s.Alarms)]
	return out
}
