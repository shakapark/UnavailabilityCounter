package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"UnavailabilityCounter/src/config"
	"UnavailabilityCounter/src/prober"
)

type collector struct {
	instances []config.Instance
}

func b2i(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (c collector) Collect(ch chan<- prometheus.Metric) {

	for _, instance := range c.instances {

		log.Debugln("Instance: " + instance.Name)
		for gName, group := range instance.Groups {
			log.Debugln("Instance: " + instance.Name + "/Group: " + gName)
			indispo := Instances.GetInstance(instance.Name).GetIndispos().GetIndispo(gName)
			if indispo.IsMaintenanceEnable() { //If Maintenance Enable For This Group
				for _, address := range group.Targets {

					ch <- prometheus.MustNewConstMetric(
						prometheus.NewDesc("probe_success_"+instance.Name+"_"+gName, "Displays whether or not the probe was a success", []string{"target"}, nil),
						prometheus.GaugeValue,
						float64(1),
						address)
				}

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("maintenance", "Displays whether or not the maintenance was enable", []string{"instance_name", "group_name"}, nil),
					prometheus.GaugeValue,
					float64(1),
					instance.Name, gName)
			} else { //If Maintenance Disable For This Group

				if group.Timeout == "" {
					group.Timeout = "10s"
				}

				//groupSuccess := 0
				for _, address := range group.Targets {

					success, err := prober.Probe(group.Kind, address, group.Timeout)
					if err != nil {
						log.Warnln("Error: ", err)
					}

					//groupSuccess += int(success)

					ch <- prometheus.MustNewConstMetric(
						prometheus.NewDesc("probe_success_"+instance.Name+"_"+gName, "Displays whether or not the probe was a success", []string{"target"}, nil),
						prometheus.GaugeValue,
						float64(b2i(success)),
						address)
				}

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc("maintenance", "Displays whether or not the maintenance was enable", []string{"instance_name", "group_name"}, nil),
					prometheus.GaugeValue,
					float64(0),
					instance.Name, gName)
			}

			//register(groupSuccess, instance.Name, gName)
		}
	}

}

func probeHandler(w http.ResponseWriter, r *http.Request, c *config.Config) {

	log.Debugln("Start Handler Function")
	registry := prometheus.NewRegistry()
	instances := c.Counter //instances:[]Instance

	log.Debugln("Lunch Collector")
	collector := collector{instances: instances}
	registry.MustRegister(collector)

	log.Debugln("Show Result")
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
