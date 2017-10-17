package main

import(
	"net"
	"net/http"
	"strings"
	
	"github.com/prometheus/common/log"
)

func probeGoogle() bool {
	
	conn, err := net.Dial("tcp4","google.fr:80")
	if err != nil {
		log.Infoln("err", err)
		return false
	}

	defer conn.Close()

	return true
}

func ProbeTCP(target string) bool {

	conn, err := net.Dial("tcp4",target)
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
		resp, err := http.Get(target)
		
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
