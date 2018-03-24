package main

import(
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	promlog "github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	
	"github.com/shakapark/UnavailabilityCounter/config"
	"github.com/shakapark/UnavailabilityCounter/indispo"
)

var(
	sc = config.SafeConfig{
		C: &config.Config{},
	}
	
	log promlog.Logger
	
	Indispos []*indispo.Indispo
	
	configFile = kingpin.Flag("config.file", "Compteur configuration file.").Default("indispo.yml").String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9143").String()
	logLevel = kingpin.Flag("log.level", "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]").Default("info").String()
)

func setInstanceName(instanceName, groupName string) string {
	return instanceName+"-"+groupName
}

func getIndispoFromName(tab []*indispo.Indispo, name string) *indispo.Indispo {
	for _, indispo := range tab {
		if indispo.GetName() == name {
			return indispo
		}
	}
	return nil
}

func setIndispos(c *config.Config) {
	instances := c.Counter
	log.Debugln("Instance Count: "+strconv.FormatInt(int64(len(instances)), 10))
	Indispos = []*indispo.Indispo{}
	for _, instance := range instances {
		for gName, _ := range instance.Groups {
			log.Debugln("Indispo Name: "+setInstanceName(instance.Name, gName))
			Indispos = append(Indispos, indispo.New(setInstanceName(instance.Name, gName)))
			log.Debugln("Indispos Count: "+strconv.FormatInt(int64(len(Indispos)), 10))
		}
	}
}

func hasMaintenancesEnable() bool {
	for _, indispo := range Indispos {
		if indispo.IsMaintenanceEnable() {
			return true
		}
	}
	return false
}

func reloadConfig(reloadCh chan<- chan error) {
	rc := make(chan error)
	reloadCh <- rc
	if err := <-rc; err != nil {
		log.Errorln("Error: failed to reload config: %s", err)
	}else{
		sc.Lock()
		conf := sc.C
		sc.Unlock()
		setIndispos(conf)
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("compteur_indispo"))
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	log = promlog.Base()
	if err := log.SetLevel(*logLevel); err != nil {
		log.Fatal("Error: ", err)
	}	

	log.Infoln("msg", "Starting Compteur Indispo", "version", version.Info())
	log.Infoln("msg", "Build context", version.BuildContext())

	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatal("Error loading config", err)
		os.Exit(1)
	}
	log.Infoln("msg", "Loaded config file")
	sc.Lock()
	conf := sc.C
	sc.Unlock()
	setIndispos(conf)
	
	for i, indispo := range Indispos {
		log.Debugln("Indispo["+strconv.FormatInt(int64(i), 10)+"]: "+indispo.GetName())
	}

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
	
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		sc.Lock()
		conf := sc.C
		sc.Unlock()
		probeHandler(w, r, conf)
	})
	
	http.HandleFunc("/api/maintenance", func(w http.ResponseWriter, r *http.Request) {
		maintenanceHandler(w, r)
	})
	
	/*http.HandleFunc("/api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
		queryRangeHandler(w, r)
	})

	http.HandleFunc("/api/v1/label/__name__/values", func(w http.ResponseWriter, r *http.Request) {
		checkAPI(w, r)
	})*/

	http.HandleFunc("/-/reload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "This endpoint requires a POST request.\n")
			return
		}

		if !hasMaintenancesEnable(){
			reloadConfig(reloadCh)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<head>
					<title>Compteur d'Indisponibilite</title>
				</head>
				<body>
					<h1>Compteur d'Indisponibilite</h1>
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
