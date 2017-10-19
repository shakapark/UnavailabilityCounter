package main

import(
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"
	//"net/smtp"
	"strings"
	"sync"
	"syscall"

	yaml "gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Indispo struct {
	Progress bool
	StartTimeStamp time.Time
	StopTimeStamp time.Time
	TimeStampBack time.Time
}

type Config struct {
	Counter map[string]Group `yaml:"count"`
	XXX map[string]interface{} `yaml:",inline"`
}

type Group struct {
	Targets []string `yaml:"targets"`
	Kind string `yaml:"kind"`
}

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

var(
	sc = SafeConfig{
		C: &Config{},
	}
	
	Maintenance bool
	maintenanceTexte string
	
	GroupNames []string
	Indispos map[string]Indispo
	
	configFile = kingpin.Flag("config.file", "Compteur configuration file.").Default("comptindispo.yml").String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9143").String()
)

func getGroupNames(c *Config){
	groups := c.Counter
	GroupNames = []string{}
	for name, _ := range groups {
		tmp := make([]string, len(GroupNames)+1, len(GroupNames)+1)
		for i, _ := range GroupNames {
			tmp[i] = GroupNames[i]
		}
		tmp[len(GroupNames)] = name
		GroupNames = tmp
		err := os.Mkdir("/data/"+name, 0777)
		if err != nil {
			log.Infoln("Msg:", err)
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

func (sc *SafeConfig) ReloadConfig(confFile string) (err error) {
	var c = &Config{}

	yamlFile, err := ioutil.ReadFile(confFile)
	if err != nil {
		return fmt.Errorf("Error reading config file: %s", err)
	}

	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		return fmt.Errorf("Error parsing config file: %s", err)
	}

	sc.Lock()
	sc.C = c
	sc.Unlock()
	
	getGroupNames(c)
	getIndispos(GroupNames)
	
	return nil
}

func checkOverflow(m map[string]interface{}, ctx string) error {
	if len(m) > 0 {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		return fmt.Errorf("unknown fields in %s: %s", ctx, strings.Join(keys, ", "))
	}
	return nil
}

func (s *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	if err := checkOverflow(s.XXX, "config"); err != nil {
		return err
	}
	return nil
}

func init() {
	prometheus.MustRegister(version.NewCollector("compteur_indispo"))
}

func probeHandler(w http.ResponseWriter, r *http.Request, c *Config) {

	registry := prometheus.NewRegistry()
	groups := c.Counter   					//groups:map[string]Group
	for groupName, group := range groups {
		
		switch group.Kind {

		case "http":
			probeSuccessGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: "probe_success_"+groupName,
					Help: "Displays whether or not the probe was a success",
				},[]string{"target"})
			registry.MustRegister(probeSuccessGauge)
			
			var t = make([]int,len(group.Targets),len(group.Targets))
			
			for i, address := range group.Targets {
				
				success := ProbeHTTP(address)
				if success {
					probeSuccessGauge.With(prometheus.Labels{"target":address}).Set(1)
					t[i] = 1
				} else {
					if Maintenance {
						probeSuccessGauge.With(prometheus.Labels{"target":address}).Set(1)
						t[i] = 1
					}else{
						probeSuccessGauge.With(prometheus.Labels{"target":address}).Set(0)
						t[i] = 0
					}
				}
			}
			var somme int
			somme = 0
			for _, v := range t {
				somme += v
			}
			register(somme, groupName)

		case "tcp":
			probeSuccessGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: "probe_success_"+groupName,
					Help: "Displays whether or not the probe was a success",
				},[]string{"target"})
			registry.MustRegister(probeSuccessGauge)
			
			var t = make([]int,len(group.Targets),len(group.Targets))
			
			for i, address := range group.Targets {
				
				success := ProbeTCP(address)
				if success {
					probeSuccessGauge.With(prometheus.Labels{"target":address}).Set(1)
					t[i] = 1
				} else {
					if Maintenance {
						probeSuccessGauge.With(prometheus.Labels{"target":address}).Set(1)
						t[i] = 1
					}else{
						probeSuccessGauge.With(prometheus.Labels{"target":address}).Set(0)
						t[i] = 0
					}
				}
			}
			var somme int
			somme = 0
			for _, v := range t {
				somme += v
			}
			register(somme, groupName)
		
		default:
			log.Infoln("err", "Unknown kind request : ", group.Kind)
		}
	}

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
		log.Fatal("msg", "Error loading config", "err", err)
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
		}

		http.Redirect(w, r, "/", 301)
	})
	
	http.HandleFunc("/api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
		queryRangeHandler(w, r, GroupNames)
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
