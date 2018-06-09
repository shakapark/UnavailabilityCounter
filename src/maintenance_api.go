package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type postJSON struct {
	Instance string
	Name     string
	Action   string
}

func maintenanceHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case "POST":
		var js []postJSON
		bodyBytes, errio := ioutil.ReadAll(r.Body)
		if errio != nil {
			log.Warnln("Error: ", errio)
			return
		}
		log.Debugln(string(bodyBytes))
		err := json.Unmarshal(bodyBytes, &js)
		if err != nil {
			log.Warnln("Error: ", err)
			return
		}

		log.Debugln(js)

		for _, j := range js {
			log.Debugln(j.Instance)
			log.Debugln(j.Name)
			instance := Instances.GetInstance(j.Instance)
			if instance == nil {
				log.Infoln("Error: Instance " + j.Instance + " don't exist")
				break
			}
			i := instance.GetIndispos().GetIndispo(j.Name)
			if i == nil {
				log.Infoln("Error: Indispo " + j.Name + " don't exist")
				break
			}

			if j.Action == "enable" {
				i.EnableMaintenance()
				log.Infoln("Maintenance has been enable for " + j.Name)
			} else if j.Action == "disable" {
				i.DisableMaintenance()
				log.Infoln("Maintenance has been disable for " + j.Name)
			} else {
				log.Infoln("Error: " + j.Action + " is not a valid action (action: [enable|disable])")
				break
			}
		}

	case "GET":
		w.Header().Set("Content-Type", "application/json")
		str, err := Instances.GetStatus()
		if err != nil {
			log.Warnln("Error: ", err)
			return
		}
		w.Write([]byte(str))

	default:
		log.Infoln("Error: ", r.Method, " is not a valid method (method: [GET|POST])")
	}
}
