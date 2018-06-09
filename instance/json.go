package instance

import (
	"UnavailabilityCounter/indispo"
	"encoding/json"
)

//JSON Json encoding of Instance
type JSON struct {
	Instance     string        `JSON:"instance"`
	JSONIndispos indispo.Jsons `JSON:"indispos"`
}

//JSONs Json encoding of Instances
type JSONs struct {
	List []JSON
}

func (i *Instance) toJSON() *JSON {
	return &JSON{
		Instance:     i.GetName(),
		JSONIndispos: *i.GetIndispos().ToJSON(),
	}
}

//GetStatus Return Instance status in json format
func (i *Instance) GetStatus() (string, error) {
	j := i.toJSON()
	msg, err := json.Marshal(j)
	return string(msg), err
}

//ToJSON Translate Instances in JSONs
func (is *Instances) ToJSON() *JSONs {
	var list []JSON
	for _, i := range is.GetList() {
		list = append(list, *i.toJSON())
	}

	return &JSONs{
		List: list,
	}
}

//GetStatus Return Instances status in json format
func (is *Instances) GetStatus() (string, error) {
	j := is.ToJSON()
	msg, err := json.Marshal(j)
	return string(msg), err
}
