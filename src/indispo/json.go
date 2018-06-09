package indispo

import (
	"encoding/json"
)

//JSONMaintenance Json encoding of maintenance
type JSONMaintenance struct {
	Status     string `json:"status"`
	LastUpdate string `json:"last-update"`
}

//JSON Json encoding of an indispo
type JSON struct {
	Name            string          `json:"name"`
	Status          string          `json:"status"`
	JSONMaintenance JSONMaintenance `json:"maintenance"`
}

//Jsons Json encoding of an indispos
type Jsons struct {
	List []JSON
}

func (i *Indispo) toJSON() *JSON {
	var status, mStatus string
	if i.Progress {
		status = "Unavailable"
	} else {
		status = "Available"
	}
	if i.getMaintenance().GetStatus() {
		mStatus = "Maintenance Enable"
	} else {
		mStatus = "Maintenance Disable"
	}

	return &JSON{
		Name:   i.GetName(),
		Status: status,
		JSONMaintenance: JSONMaintenance{
			Status:     mStatus,
			LastUpdate: i.getMaintenance().GetLastUpdate().String(),
		},
	}
}

//GetStatus Return the indispo's status in json
func (i *Indispo) GetStatus() (string, error) {
	j := i.toJSON()
	msg, err := json.Marshal(j)
	return string(msg), err
}

//ToJSON Return Jsons pointer corresponding to Indispos
func (is *Indispos) ToJSON() *Jsons {
	var list []JSON
	for _, i := range is.GetList() {
		list = append(list, *i.toJSON())
	}

	return &Jsons{
		List: list,
	}
}

//GetStatus Return the indispos status in json
func (is *Indispos) GetStatus() (string, error) {
	j := is.ToJSON()
	msg, err := json.Marshal(j)
	return string(msg), err
}
