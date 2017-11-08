package main

import(
	"context"
	"net"
	"net/http"
	"strings"
	"time"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type collector struct {
	instances []Instance
}

func probeGoogle() bool {
	
	d, _ := time.ParseDuration("1s")
	conn, err := net.DialTimeout("tcp4","google.fr:80", d)
	if err != nil {
		log.Infoln("err", err)
		return false
	}

	defer conn.Close()

	return true
}

func ProbeTCP(target string) bool {
	
	d, _ := time.ParseDuration("1s")
	conn, err := net.DialTimeout("tcp4",target, d)
	if err != nil {
		success := probeGoogle()
		if success == true {
			return false
		}else{
			return true
		}
	}

	defer conn.Close()

	return true
}

func ProbeHTTP(target string) bool {
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "http://" + target
	}
	
	var code int
	code = 305
	
	for code >= 301 && code <= 308 {
		d, _ := time.ParseDuration("1s")
		ctx, _ := context.WithTimeout(context.Background(), d)
		req, err := http.NewRequest("GET", target, nil)
		req = req.WithContext(ctx)
		
		var client http.Client
		resp, err := client.Do(req)

		if err != nil {
			success := probeGoogle()
			if success == true {
				return false
			}else{
				return true
			}
		}else{
			code = resp.StatusCode
			if code == 200 {
				return true
			}else if !(code >= 301 && code <= 308) {
				success := probeGoogle()
				if success == true {
					return false
				}else{
					return true
				}
			}else{
				target = resp.Header.Get("Location")
				if err != nil {
					log.Infoln("err", err)
					return false
				}
			}
		}
		
		defer resp.Body.Close()
	}
	
	return false
}

func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (c collector) Collect(ch chan<- prometheus.Metric){
	
	for _, instance := range c.instances {
		var t2 = make([]int,len(instance.Groups),len(instance.Groups))
		var k int
		k = 0
		
		for groupName, group := range instance.Groups {
			var t = make([]int,len(group.Targets),len(group.Targets))
			var somme int
			
			switch group.Kind {

				case "http":
			
					for i, address := range group.Targets {
				
						success := ProbeHTTP(address)
						if success {
							t[i] = 1
						} else {
							if Maintenance {
								t[i] = 1
							}else{
								t[i] = 0
							}
						}
				
						ch <- prometheus.MustNewConstMetric(
						  prometheus.NewDesc("probe_success_"+instance.Name+"_"+groupName, "Displays whether or not the probe was a success", []string{"target"}, nil),
						  prometheus.GaugeValue,
						  float64(t[i]),
						  address)
					}
					somme = 0
					for _, v := range t {
						somme += v
					}
					register(somme, instance.Name, groupName)

				case "tcp":

					for i, address := range group.Targets {
				
						success := ProbeTCP(address)
						if success {
							t[i] = 1
						} else {
							if Maintenance {
								t[i] = 1
							}else{
								t[i] = 0
							}
						}
				
						ch <- prometheus.MustNewConstMetric(
						  prometheus.NewDesc("probe_success_"+instance.Name+"_"+groupName, "Displays whether or not the probe was a success", []string{"target"}, nil),
						  prometheus.GaugeValue,
						  float64(t[i]),
						  address)
					}
					somme = 0
					for _, v := range t {
						somme += v
					}
					register(somme, instance.Name, groupName)
		
				default:
					log.Infoln("err", "Unknown kind request : ", group.Kind)
			}
			t2[k]=somme
			k+=1
		}
		
		var somme2 bool
		somme2 = false
		for _, v := range t2 {
			if v == 0 {
				somme2 = true
			}
		}
		
		registerG(somme2, instance.Name)
	}
}
