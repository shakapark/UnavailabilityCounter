package indispo

import(
	"encoding/json"
	"time"

	"github.com/shakapark/UnavailabilityCounter/maintenance"
)

type Json struct {
    Name        string `json:"name"`
	Status      string `json:"status"`
	JsonMaintenance struct {
		Status     string `json:"status"`
		LastUpdate string `json:"last-update"`
	} `json:"maintenance"`
}

type Indispo struct {
	Name           string
	Progress       bool
	StartTimeStamp time.Time
	StopTimeStamp  time.Time
	TimeStampBack  time.Time
	Maintenance    maintenance.Maintenance
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
	return &i.Maintenance
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
	return &Json{
		Name:   i.GetName(),
		Status: i.GetStatus(),
		JsonMaintenance: JsonMaintenance{
			Status:     i.getMaintenance().GetStatus(),
			LastUpdate: i.getMaintenance().GetLastUpdate().String(),
		},
	}
}

func (i *Indispo) GetStatus() (string, error) {
	j := i.toJson()
	return json.Marshal(j)
}
