package grafana

import (
	"UnavailabilityCounter/src/instance"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type check struct {
	Status string   `json:"status"`
	Data   []string `json:"data,omitempty"`
}

func checkAPI(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var c check
	c.Status = "success"
	c.Data = []string{"counter", "time"}
	content, err := json.Marshal(c)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Write(content)

}

func queryRangeHandler(w http.ResponseWriter, r *http.Request, Instances *instance.Instances) {

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "'query' parameter must be specified", 400)
		return
	}

	start, err := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	end, err := strconv.ParseInt(r.URL.Query().Get("end"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	step, err := strconv.ParseInt(r.URL.Query().Get("step"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	//Check Query Type
	tmp := strings.Split(query, "{")

	if tmp[0] != "counter" && tmp[0] != "time" {
		http.Error(w, "Unknown request : "+tmp[0], 400)
		return
	}
	query = tmp[0]

	var sendGrafana sendData

	switch query {

	case "time":
		if len(tmp) > 1 {
			tmp2 := strings.Split(tmp[1], "}")
			tmp3 := strings.Split(tmp2[0], ",")

			tmp4 := strings.Split(tmp3[0], "=")
			if tmp4[0] != "instance" {
				http.Error(w, "Unknown parameter : "+tmp4[0], 400)
				return
			}
			tmp5 := strings.Split(tmp4[1], "\"")
			instance := tmp5[0]

			switch len(tmp3) {

			case 1:
				sendGrafana = queryTime(start, end, step, query, instance, "", Instances)

			case 2:
				tmp6 := strings.Split(tmp3[1], "=")
				if tmp6[0] != "group" {
					http.Error(w, "Unknown parameter : "+tmp6[0], 400)
					return
				}
				tmp7 := strings.Split(tmp6[1], "\"")
				group := tmp7[0]
				sendGrafana = queryTime(start, end, step, query, instance, group, Instances)

			default:
				http.Error(w, "Error parsing query", 400)
				return
			}
		} else {
			http.Error(w, "Instance must be specified", 400)
			return
		}

	default:
		http.Error(w, "Unknown request : "+tmp[0], 400)
		return

	}

	sendGrafana.Status = "success"

	w.Header().Set("Content-Type", "application/json")
	content, err := json.Marshal(sendGrafana)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Write(content)
}
