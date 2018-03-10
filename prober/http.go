package prober

import(
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

func ProbeHTTP(target string, timeout string) (bool, error) {
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "http://" + target
	}
	
	var code int
	code = 305
	
	d, r := time.ParseDuration(timeout)
	if r != nil {
		return false, r
	}
	
	ctx, _ := context.WithTimeout(context.Background(), d)
	
	for code >= 301 && code <= 308 {

		req, err := http.NewRequest("GET", target, nil)
		req = req.WithContext(ctx)
		
		var client http.Client
		resp, err := client.Do(req)

		if err != nil {
			success, err2 := probeGoogle(timeout)
			return !success, err2
		}else{
			code = resp.StatusCode
			if code == 200 {
				return true, nil
			}else if !(code >= 301 && code <= 308) {
				success, err2 := probeGoogle(timeout)
				return !success, err2
			}else{
				target = resp.Header.Get("Location")
				if err != nil {
					return false, err
				}
			}
		}
		
		defer resp.Body.Close()
	}
	
	return false, nil
}

