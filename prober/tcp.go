package prober

import(
	"errors"
	"net"
	"time"
)

func probeTCP(target string, timeout string) (bool, error) {
	
	d, r := time.ParseDuration(timeout)
	if r != nil {
		return false, r
	}
	conn, err := net.DialTimeout("tcp4", target, d)
	defer conn.Close()
	
	if err != nil {
		success, err2 := probeGoogle(timeout)
		return !success, err2
	}

	return true, nil
}

