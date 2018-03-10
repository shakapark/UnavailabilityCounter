package indispo

import(
	"encoding/json"
	"time"

	"github.com/shakapark/UnavailabilityCounter/maintenance"
)

type JsonMaintenance struct {
	Status     string `json:"status"`
	LastUpdate string `json:"last-update"`
}

type Json struct {
    Name            string          `json:"name"`
	Status          string          `json:"status"`
	JsonMaintenance JsonMaintenance `json:"maintenance"`
}

type Indispo struct {
	Name           string
	Progress       bool
	StartTimeStamp time.Time
	StopTimeStamp  time.Time
	TimeStampBack  time.Time
	Maintenance    *maintenance.Maintenance
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

func (i *Indispo) toJson() *Json {
	if i.Progress {
		status := "Unavailable"
	}else{
		status := "Available"
	}
	if i.getMaintenance().GetStatus() {
		mStatus := "Maintenance Enable"
	}else{
		mStatus := "Maintenance Disable"
	}
	status := i.Progress
	return &Json{
		Name:   i.GetName(),
		Status: status,
		JsonMaintenance{
			Status:     mStatus,
			LastUpdate: i.getMaintenance().GetLastUpdate().String(),
		},
	}
}

func (i *Indispo) GetStatus() (string, error) {
	j := i.toJson()
	msg, err := json.Marshal(j)
	return string(msg), err
}
