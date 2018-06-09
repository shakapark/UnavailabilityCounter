package indispo

import (
	"time"

	"UnavailabilityCounter/maintenance"
)

//Indispo Represent Unavailability of a target
type Indispo struct {
	Name           string
	Progress       bool
	StartTimeStamp time.Time
	StopTimeStamp  time.Time
	TimeStampBack  time.Time
	Maintenance    *maintenance.Maintenance
}

//Indispos Represent a list of Indispo
type Indispos struct {
	list []*Indispo
}

//New Return an Indispo with the its name
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

//Start Begin the mesure of unavailability
func (i *Indispo) Start() {
	if !i.Progress {
		i.StartTimeStamp = time.Now()
		i.Progress = true
	}
}

//Stop End the mesure of unavailability
func (i *Indispo) Stop() {
	if i.Progress {
		i.StopTimeStamp = time.Now()
		i.Progress = false
	}
}

//GetName Return indispo's name
func (i *Indispo) GetName() string {
	return i.Name
}

//IsProgress Return true if indispo is recording
func (i *Indispo) IsProgress() bool {
	return i.Progress
}

func (i *Indispo) getMaintenance() *maintenance.Maintenance {
	return i.Maintenance
}

//IsMaintenanceEnable Return true if indipo's maintenance is enable
func (i *Indispo) IsMaintenanceEnable() bool {
	return i.getMaintenance().GetStatus()
}

//EnableMaintenance Enable indispo's maintenance
func (i *Indispo) EnableMaintenance() {
	i.getMaintenance().Enable()
}

//DisableMaintenance Disable indispo's maintenance
func (i *Indispo) DisableMaintenance() {
	i.getMaintenance().Disable()
}

//News Return an empty list of indispo
func News() *Indispos {
	return &Indispos{
		list: []*Indispo{},
	}
}

//Add Add an indispo in the list
func (is *Indispos) Add(name string) {
	is.list = append(is.list, New(name))
}

//GetList Return indispo array
func (is *Indispos) GetList() []*Indispo {
	return is.list
}

//IsProgress Return true if one of indispo is recording
func (is *Indispos) IsProgress() bool {
	for _, i := range is.list {
		if i.Progress {
			return true
		}
	}
	return false
}

//GetIndispo Return the indispo with the name in parameter
//(nil if the name isn't in the list)
func (is *Indispos) GetIndispo(name string) *Indispo {
	for _, i := range is.list {
		if i.GetName() == name {
			return i
		}
	}
	return nil
}

//HasMaintenancesEnable Return true if one indispo in the list
// has its maintenance enabled
func (is *Indispos) HasMaintenancesEnable() bool {
	for _, i := range is.list {
		if i.IsMaintenanceEnable() {
			return true
		}
	}
	return false
}
