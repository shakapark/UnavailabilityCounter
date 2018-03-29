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

type Jsons struct {
	List []Json
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

func (is *Indispos) ToJson() *Jsons {
	var list []Json
	for _, i := range is.GetList() {
		list = append(list, *i.toJson())
	}
	
	return &Jsons{
		list: list,
	}
}

func (is *Indispos) GetStatus() (string, error) {
	j := is.ToJson()
	msg, err := json.Marshal(j)
	return string(msg), err
}
