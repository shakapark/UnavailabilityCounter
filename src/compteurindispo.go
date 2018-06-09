package main

import (
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

	"UnavailabilityCounter/src/config"
	"UnavailabilityCounter/src/instance"
)

var (
	sc = config.SafeConfig{
		C: &config.Config{},
	}

	log promlog.Logger

	//Instances Variables that contains the list of instance
	Instances *instance.Instances

	configFile    = kingpin.Flag("config.file", "Compteur configuration file.").Default("indispo.yml").String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9143").String()
	logLevel      = kingpin.Flag("log.level", "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]").Default("info").String()
)

func setInstances(c *config.Config) {
	log.Debugln("Instance Count: " + strconv.FormatInt(int64(len(c.Counter)), 10))
	Instances = instance.News()
	for _, counter := range c.Counter {
		Instances.Add(counter.Name)
		for gName := range counter.Groups {
			Instances.GetInstance(counter.Name).AddIndispo(gName)
		}
	}
}

func reloadConfig(reloadCh chan<- chan error) {
	rc := make(chan error)
	reloadCh <- rc
	if err := <-rc; err != nil {
		log.Errorln("Error: failed to reload config: %s", err)
	} else {
		sc.Lock()
		conf := sc.C
		sc.Unlock()
		setInstances(conf)
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("unavailabilitycounter"))
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	log = promlog.Base()
	if err := log.SetLevel(*logLevel); err != nil {
		log.Fatal("Error: ", err)
	}

	log.Infoln("Msg", "Starting UnavailabilityCounter")

	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatal("Error loading config", err)
		os.Exit(1)
	}
	log.Infoln("Msg", "Loaded config file")
	sc.Lock()
	conf := sc.C
	sc.Unlock()
	setInstances(conf)

	hup := make(chan os.Signal)
	reloadCh := make(chan chan error)
	signal.Notify(hup, syscall.SIGHUP)

	go func() {
		for {
			select {
			case <-hup:
				if err := sc.ReloadConfig(*configFile); err != nil {
					log.Infoln("Msg", "Error reloading config", "err", err)
					continue
				}
				log.Infoln("Msg", "Reloaded config file")
			case rc := <-reloadCh:
				if err := sc.ReloadConfig(*configFile); err != nil {
					log.Infoln("Msg", "Error reloading config", "err", err)
					rc <- err
				} else {
					log.Infoln("Msg", "Reloaded config file")
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

		for _, i := range Instances.GetList() {
			if i.GetIndispos().IsProgress() {
				log.Warnln("Error, you can't reload config because unaivability is in progress.")
				fmt.Fprintf(w, "Error, you can't reload config because unaivability is in progress.\n")
				return
			}
		}
		reloadConfig(reloadCh)
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
				</body>
			</html>`))
	})

	log.Infoln("msg", "Listening on", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
