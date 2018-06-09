package prober

import (
	"net"
	"time"
)

func probeTCP(target string, timeout string) (bool, error) {

	d, r := time.ParseDuration(timeout)
	if r != nil {
		return false, r
	}

	conn, err := net.DialTimeout("tcp4", target, d)
	if err != nil {
		success, err2 := probeGoogle(timeout)
		return !success, err2
	}

	defer conn.Close()

	return true, nil
}
