package instance

import(
	"encoding/json"
	
	"github.com/shakapark/UnavailabilityCounter/indispo"
)

type Json struct {
    Instance		string           `json:"instance"`
	JsonIndispos	indispo.Jsons	 `json:"indispos"`
}

type Jsons struct {
	list []Json
}

func (i *Instance) toJson() *Json {
	return &Json{
		Instance:     i.GetName(),
		JsonIndispos: i.GetIndispos.ToJson(),
	}
}

func (i *Instance) GetStatus() (string, error) {
	j := i.toJson()
	msg, err := json.Marshal(j)
	return string(msg), err
}

func (is *Instances) ToJson() *Jsons {
	var list []Json
	for _, i := range is.GetList() {
		list = append(list, *i.toJson())
	}
	
	return &Jsons{
		list: list,
	}
}

func (is *Instances) GetStatus() (string, error) {
	j := is.ToJson()
	msg, err := json.Marshal(j)
	return string(msg), err
}
