package grafana

import (
	"UnavailabilityCounter/src/instance"
)

func getPercent(s int64, t int64) float64 {
	var result float64
	result = float64(s) * 100 / float64(t)
	result = 100 - result
	return result
}

func toString(tab []string) string {
	var r string
	for _, s := range tab {
		r += s + "|"
	}
	return r
}

// func contains(t string, tab []string) bool {
// 	for _, tmp := range tab {
// 		if tmp == t {
// 			return true
// 		}
// 	}
// 	return false
// }

func contains(t string, instances *instance.Instances) bool {
	tab := instances.GetList()
	for _, tmp := range tab {
		if tmp.GetName() == t {
			return true
		}
	}
	return false
}
