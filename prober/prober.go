package prober

import(
	"errors"
	"time"
)

func probeGoogle(timeout string) (bool, error) {
	
	d, r := time.ParseDuration(timeout)
	if r != nil {
		return false, r
	}
	conn, err := net.DialTimeout("tcp4","google.fr:80", d)
	if err != nil {
		return false, err
	}

	defer conn.Close()

	return true
}

func Probe(kind, target, timeout string) (bool, error) {
	
	switch kind {
		case "tcp":
			return probeTCP(target, timeout)
		case "http":
			return probeHTTP(target, timeout)
		default:
			return false, errors.New("Unknow request: " + kind)
	}	
}
