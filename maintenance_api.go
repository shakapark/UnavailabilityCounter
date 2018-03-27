package main

import (
	"net/http"
)

func maintenanceHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		query := r.URL.Query().Get("indispo")
		if query == "" {
			http.Error(w, "indispo parameter must be specified", 400)
			return
		}
		
		for _, indispo := range Indispos.GetList() {
			
			if indispo.GetName() == query {
				
				if indispo.IsMaintenanceEnable() {
					indispo.DisableMaintenance()
				}else{
					indispo.EnableMaintenance()
				}
				http.Redirect(w, r, "/api/maintenance", 301)
				return
			}
		}
		
		log.Warnln("Error: "+query+" not found")
		http.Redirect(w, r, "/api/maintenance", 301)
		return
	}

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		str, err := Indispos.GetStatus()
		if err != nil {
			log.Warnln("Error: ", err)
		}		
		w.Write([]byte(str))
	}
}
