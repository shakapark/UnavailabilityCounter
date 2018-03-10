package indispo

import(
	"encoding/json"
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

func (i *Indispo) toJson() *Json {
	var status, mStatus string
	if i.Progress {
		status = "Unavailable"
	}else{
		status = "Available"
	}
	if i.getMaintenance().GetStatus() {
		mStatus = "Maintenance Enable"
	}else{
		mStatus = "Maintenance Disable"
	}
	
	return &Json{
		Name:            i.GetName(),
		Status:          status,
		JsonMaintenance: JsonMaintenance{
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
