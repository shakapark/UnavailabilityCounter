package indispo

import(
	"time"

	"github.com/shakapark/UnavailabilityCounter/maintenance"
)

type Indispo struct {
	Name           string
	Progress       bool
	StartTimeStamp time.Time
	StopTimeStamp  time.Time
	TimeStampBack  time.Time
	Maintenance    *maintenance.Maintenance
}

type Indispos struct {
	List []*Indispo
}

func New(name string) *Indispo {
	t := time.Now()
	return &Indispo{
		Name:           name,
		Progress:       false,
		StartTimeStamp: t,
		StopTimeStamp:  t,
		TimeStampBack:  t,
		Maintenance:    maintenance.New(),
	}
}

func (i *Indispo) Start() {
	if !i.Progress {
		i.StartTimeStamp = time.Now()
		i.Progress = true
	}
}

func (i *Indispo) Stop() {
	if i.Progress {
		i.StopTimeStamp = time.Now()
		i.Progress = false
	}
}

func (i *Indispo) GetName() string {
	return i.Name
}

func (i *Indispo) getMaintenance() *maintenance.Maintenance {
	return i.Maintenance
}

func (i *Indispo) IsMaintenanceEnable() bool {
	return i.getMaintenance().IsEnable()
}

func (i *Indispo) EnableMaintenance() {
	i.getMaintenance().Enable()
}

func (i *Indispo) DisableMaintenance() {
	i.getMaintenance().Disable()
}

func (is *Indispos) GetIndispo(name string) *Indispo {
	for _, i := range is.List {
		if i.GetName() == name {
			return i
		}
	}
	return nil
}

func (is *Indispos) HasMaintenancesEnable() bool {
	for _, i := range is.List {
		if i.IsMaintenanceEnable() {
			return true
		}
	}
	return false
}
