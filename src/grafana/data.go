package grafana

import (
	"UnavailabilityCounter/src/instance"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"

	ucdata "UnavailabilityCounter/src/data"
)

func addValue(a []model.SamplePair, n model.Time, v model.SampleValue) []model.SamplePair {
	var t = make([]model.SamplePair, len(a)+1)
	for i, v := range a {
		t[i] = v
	}
	var m model.SamplePair
	m.Timestamp = n
	m.Value = v
	t[len(a)] = m
	return t
}

type sendData struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	ResultType string   `json:"resultType"`
	Results    []result `json:"result"`
}

type result struct {
	Metric metric             `json:"metric"`
	Values []model.SamplePair `json:"values"`
}

type metric struct {
	Name     string `json:"__name__"`
	Job      string `json:"job"`
	Instance string `json:"instance"`
	Group    string `json:"instance"`
}

func queryTimeByMonth(year int, month string, instance string, group string, start int64, end int64, step int64) (int64, error) { //month:now|January|February|...|November|December

	if start >= end {
		log.Infoln("err", "Bad TimeStamp, start > end!")
		return 0, nil
	}

	if time.Unix(start, 0).Month().String() != month {
		log.Infoln("err", "Bad Month")
		return 0, nil
	}

	var sum int64
	sum = 0

	var path string
	if !(contains(group, GroupNames)) {
		if group == "" {
			path = "/data/" + instance + "/" + month + strconv.Itoa(year)
		} else {
			return 0, errors.New("Unknown group " + group)
		}
	} else {
		path = "/data/" + instance + "/" + group + "/" + month + strconv.Itoa(year)
	}

	contentFile, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var saveData ucdata.Data
	if err := json.Unmarshal(contentFile, &saveData); err != nil {
		return 0, err
	}

	for start <= end {
		for _, i := range saveData.Data {
			if start <= i[0] { //i[0] > start
				if end >= i[1] {
					sum += i[1] - i[0]
					start = i[1]
				} else if end > i[0] {
					sum += end - i[0]
					start = i[1]
				}
			} else if start <= i[1] { //i[0] < start < i[1]
				if end < i[1] {
					sum += end - start
				} else {
					sum += i[1] - start
				}
				start = i[1]
			}
		}
		start += step
	}

	return sum, nil
}

func queryTime(start int64, end int64, step int64, query string, instance string, group string, Instances *instance.Instances) sendData {
	results := make([]result, 1)
	timeTotal := end - start
	var result result
	result.Metric = metric{Name: query, Job: "compteur", Instance: instance, Group: group}
	var v []model.SamplePair
	var sum int64
	var p float64
	var monthS string
	var year int

	for start <= end {

		year = time.Unix(start, 0).Year()
		monthS = time.Unix(start, 0).Month().String()

		if !(contains(instance, Instances)) {
			log.Infoln("Error : Unknown instance ", instance)
			break
		}

		endTmp := start + step
		if endTmp >= end {
			s, err := queryTimeByMonth(year, monthS, instance, group, start, end, step)
			if err != nil {
				log.Infoln("err : ", err)
			}
			sum += s
		} else {
			s, err := queryTimeByMonth(year, monthS, instance, group, start, start+step, step)
			if err != nil {
				log.Infoln("err : ", err)
			}
			sum += s
		}

		p = getPercent(sum, timeTotal)
		v = addValue(v, model.TimeFromUnix(start), model.SampleValue(p))

		start += step
	}

	result.Values = v
	results[0] = result
	return sendData{Status: "loading", Data: data{ResultType: "matrix", Results: results}}
}
