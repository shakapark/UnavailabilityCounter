package main

import(
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	//"net/smtp"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	
	"github.com/shakapark/UnavailabilityCounter/config"
)

type Indispo struct {
	Progress bool
	StartTimeStamp time.Time
	StopTimeStamp time.Time
	TimeStampBack time.Time
}

var(
	sc = &config.SafeConfig{C: &config.Config{},}
	
	Maintenance bool
	maintenanceTexte string
	
	InstancesNames []string
	GroupNames []string
	Indispos map[string]Indispo
	
	configFile = kingpin.Flag("config.file", "Compteur configuration file.").Default("comptindispo.yml").String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9143").String()
)

func getGroupNames(c *config.Config){
	instances := c.Counter			//instances: []Instance
	InstancesNames = []string{}
	GroupNames = []string{}
	
	for _, instance := range instances {
		tmp := make([]string, len(InstancesNames)+1, len(InstancesNames)+1)
			for i, _ := range InstancesNames {
				tmp[i] = InstancesNames[i]
			}
			tmp[len(InstancesNames)] = instance.Name
			InstancesNames = tmp
			err := os.Mkdir("/data/"+instance.Name, 0777)
			if err != nil {
				log.Infoln("msg:", err)
			}
		
		groups := instance.Groups
		for name, _ := range groups {
			tmp2 := make([]string, len(GroupNames)+1, len(GroupNames)+1)
			for i, _ := range GroupNames {
				tmp2[i] = GroupNames[i]
			}
			tmp2[len(GroupNames)] = name
			GroupNames = tmp2
			err := os.Mkdir("/data/"+instance.Name+"/"+name, 0777)
			if err != nil {
				log.Infoln("msg:", err)
			}
		}
	}
}

func getIndispos(ns []string){
	Indispos = make(map[string]Indispo)

	for _, n := range ns {
		var tmp Indispo
		tmp.Progress = false
		tmp.StartTimeStamp = time.Now()
		tmp.StopTimeStamp = time.Now()
		tmp.TimeStampBack = time.Now()
		Indispos[n] = tmp
	}
}

func addIndispos(ns []string){
	for _, n := range ns {
		var tmp Indispo
		tmp.Progress = false
		tmp.StartTimeStamp = time.Now()
		tmp.StopTimeStamp = time.Now()
		tmp.TimeStampBack = time.Now()
		Indispos[n] = tmp
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("compteur_indispo"))
}

func probeHandler(w http.ResponseWriter, r *http.Request, c *config.Config) {

	registry := prometheus.NewRegistry()
	instances := c.Counter   				//instances:[]Instance
	
	collector := collector{instances: instances}
	registry.MustRegister(collector)

	if Maintenance {
		probeMaintenance := prometheus.NewGauge(prometheus.GaugeOpts{
					Name: "probe_maintenance",
					Help: "Displays whether or not there is a Maintenance",
				})
			registry.MustRegister(probeMaintenance)
		probeMaintenance.Set(1)
		
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("compteur_indispo"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	maintenanceTexte = ""

	log.Infoln("msg", "Starting Compteur Indispo", "version", version.Info())
	log.Infoln("msg", "Build context", version.BuildContext())

	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatal("Error loading config", err)
		os.Exit(1)
	}
	
	log.Infoln("msg", "Loaded config file")

	hup := make(chan os.Signal)
	reloadCh := make(chan chan error)
	signal.Notify(hup, syscall.SIGHUP)
	
	go func() {
		for {
			select {
			case <-hup:
				if err := sc.ReloadConfig(*configFile); err != nil {
					log.Infoln("msg", "Error reloading config", "err", err)
					continue
				}
				log.Infoln("msg", "Reloaded config file")
			case rc := <-reloadCh:
				if err := sc.ReloadConfig(*configFile); err != nil {
					log.Infoln("msg", "Error reloading config", "err", err)
					rc <- err
				} else {
					log.Infoln("msg", "Reloaded config file")
					rc <- nil
				}
			}
		}
	}()
	
	http.HandleFunc("/-/reload",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(w, "This endpoint requires a POST request.\n")
				return
			}

			rc := make(chan error)
			reloadCh <- rc
			if err := <-rc; err != nil {
				http.Error(w, fmt.Sprintf("failed to reload config: %s", err), http.StatusInternalServerError)
			}
			
			getGroupNames(sc.C)
			getIndispos(GroupNames)
			addIndispos(InstancesNames)
		})
	
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		sc.Lock()
		conf := sc.C
		sc.Unlock()
		probeHandler(w, r, conf)
	})
	
	http.HandleFunc("/api/maintenance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			if Maintenance == false {
				maintenanceTexte = "Maintenance"
				log.Infoln("info", "Go to Maintenance")
				Maintenance = true
			}else{
				maintenanceTexte = ""
				log.Infoln("info", "Go to Normale Mode")
				Maintenance = false
			}
			http.Redirect(w, r, "/", 301)
		}

		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/html")
			if Maintenance == false {
				w.Write([]byte("Off"))
			}else{
				w.Write([]byte("On"))
			}
		}
	})
	
	http.HandleFunc("/api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
		queryRangeHandler(w, r)
	})

	http.HandleFunc("/api/v1/label/__name__/values", func(w http.ResponseWriter, r *http.Request) {
		checkAPI(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<head>
					<title>Compteur d'Indisponibilite</title>
				</head>
				<body>
					<h1>Compteur d'Indisponibilite ` + maintenanceTexte + `</h1>
					<p><a href="/probe">Probe</a></p>
					<p>
						<form method="POST" action="/api/maintenance">
							<input type="submit" value="MAINTENANCE">
						</form>
					</p>
				</body>
			</html>`))
	})

	log.Infoln("msg", "Listening on", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
