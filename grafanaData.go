package main

import(
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	"strings"
	"strconv"
	
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
)

type sendData struct {
	Status string `json:"status"`
	Data Data `json:"data"`
}

type Data struct {
	ResultType string `json:"resultType"`
	Results []Result `json:"result"`
}

type Result struct {
	Metric Metric `json:"metric"`
	Values []model.SamplePair `json:"values"`
}

type Metric struct {
	Name string `json:"__name__"`
	Job string `json:"job"`
	Instance string `json:"instance"`
}

func getPercent(s int64,t int64) float64 {
	var result float64
	result = float64(s)*100/float64(t)
	result = 100-result
	return result
}

func toString(tab []string) string {
	var r string
	for _, s := range tab {
		r += s+"|"
	}
	return r
}

func contains(t string, tab []string) bool {
	for _, tmp := range tab {
		if tmp == t {
			return true
		}
	}
	return false
}

//Add Timestamp/Value to a Tab
func addValue(a []model.SamplePair, n model.Time, v model.SampleValue) []model.SamplePair{	
	var t = make([]model.SamplePair,len(a)+1)
	for i, v := range a {
		t[i] = v
	}
	var m model.SamplePair
	m.Timestamp = n
	m.Value = v
	t[len(a)] = m
	return t
}

func queryTimeByMonth(year int, month string, instance string, start int64, end int64, step int64) (int64, error) { //month:now|January|February|...|November|December

	if start >=end {
		log.Infoln("err", "Bad TimeStamp, start > end!")
		return 0, nil
	}
	
	if time.Unix(start, 0).Month().String() != month {
		log.Infoln("err", "Bad Month")
		return 0, nil
	}
	
	var sum int64
	sum = 0
	var path = "/data/"+instance+"/"+month+strconv.Itoa(year)
	
	contentFile, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	
	var saveData SaveData
	if err := json.Unmarshal(contentFile, &saveData); err != nil {
		return 0, err
	}
	
	for start <= end {
		for _, i := range saveData.Data {
			if start <= i[0]{			//i[0] > start
				if end >= i[1]{
					sum += i[1]-i[0]
					start = i[1]
				}else if end > i[0] {
					sum += end - i[0]
					start = i[1]
				}
			}else if start <= i[1] {		//i[0] < start < i[1]
				if end < i[1] {
					sum += end - start
				}else{
					sum += i[1]-start
				}
				start = i[1]
			}
		}
		start += step
	}
	
	return sum, nil
}

func queryTime(start int64, end int64, step int64, query string, instances []string) sendData {
	results := make([]Result, 1)
	timeTotal := end-start
	var result Result
	result.Metric = Metric{Name: query, Job: "compteur", Instance: toString(instances)}
	var v []model.SamplePair
	var sum int64 = 0
	var p float64
	var monthS string
	var year int
		
	for start <= end {
		
		year = time.Unix(start, 0).Year()
		monthS = time.Unix(start, 0).Month().String()
		
		for _, instance := range instances {
			if !(contains(instance, GroupNames)) {
				log.Infoln("Error : Unknown instance ", instance)
				break
			}
			
			endTmp := start+step
			if endTmp >= end {
				s, err := queryTimeByMonth(year, monthS, instance, start, end, step)
				if err != nil {
					log.Infoln("err : ", err)
				}
				sum += s
			}else{
				s, err := queryTimeByMonth(year, monthS, instance, start, start+step, step)
				if err != nil {
					log.Infoln("err : ", err)
				}
				sum += s
			}
			

		}
		
		p = getPercent(sum, timeTotal)
		v = addValue(v, model.TimeFromUnix(start), model.SampleValue(p))
		
		start+=step
	}
	
	result.Values = v
	results[0] = result
	return sendData{Status: "loading", Data: Data{ResultType: "matrix", Results: results}}
}

func queryCounter(start int64, end int64, step int64, query string, instances []string) sendData {
	results := make([]Result, len(instances))
	for i, instance := range instances {
		if !(contains(instance, GroupNames)) {
			log.Infoln("Error : Unknown instance ", instance)
			break
		}
		var result Result
		var path string
		path = "/data/"+instance+"/"
		result.Metric = Metric{Name: query, Job: "compteur", Instance: instance}
		var v []model.SamplePair

		for start <= end {
			var contentFile []byte
			var err error
			year := time.Unix(start, 0).Year()
			month := time.Unix(start, 0).Month()
			contentFile, err = ioutil.ReadFile(path+month.String()+strconv.Itoa(year))
			
			if err != nil {
				log.Infoln("err : ", err)
			
				for time.Unix(start, 0).Month() == month {
					start += step
				}
			}else{
		
				if contentFile == nil || len(contentFile) == 0 {
					for time.Unix(start, 0).Month() == month {
						start += step
					}
				}else{
			
					var saveData SaveData
		
					if err := json.Unmarshal(contentFile, &saveData); err != nil {
						log.Fatal("err", err)
					}
		
					//Lecture de saveData.Data pour check timestamp
					for _, i := range saveData.Data {			//i => []int64
					
						if i[0] <= start {
							for start < i[1] {
								v = addValue(v, model.TimeFromUnix(start), model.SampleValue(1))
								start += step
							}
						}else{
							for start >= i[0] {
								for start <= end {
									v = addValue(v, model.TimeFromUnix(start), model.SampleValue(0))
									start += step
								}
							}
						}
					}
				
					v = addValue(v, model.TimeFromUnix(start), model.SampleValue(0))
					start += step
				}
			}
		}
		if v == nil {
			v = addValue(v, model.TimeFromUnix(start), model.SampleValue(0))
		}
		result.Values = v
		
		results[i] = result
	}
	
	return sendData{Status: "loading", Data: Data{ResultType: "matrix", Results: results}}
}

func queryRangeHandler(w http.ResponseWriter, r *http.Request, listGroup []string) {

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "'query' parameter must be specified", 400)
		return
	}

	start, err := strconv.ParseInt(r.URL.Query().Get("start"),10 , 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	
	end, err := strconv.ParseInt(r.URL.Query().Get("end"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	
	step, err := strconv.ParseInt(r.URL.Query().Get("step"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	
	//Check Query Type
	tmp := strings.Split(query, "{")

	if tmp[0] != "counter" && tmp[0] != "time" {
		http.Error(w, "Unknown request : "+tmp[0], 400)
		return
	}
	query = tmp[0]
	
	var sendGrafana sendData
	
	switch query {
	
		case "counter":
			if len(tmp) > 1 {
				tmp2 := strings.Split(tmp[1], ":")
				if len(tmp2) != 2 {
					http.Error(w, "Error parsing query", 400)
					return
				}

				if tmp2[0] != "instance" {
					http.Error(w, "Unknown parameter : "+tmp2[0], 400)
					return
				}
		
				tmp3 := strings.Split(tmp2[1], "}")
				tmp4 := strings.Split(tmp3[0], "\"")
				instance := tmp4[1]
		
				sendGrafana = queryCounter(start, end, step, query, []string{instance})
		
			}else{
				sendGrafana = queryCounter(start, end, step, query, listGroup)
			}
		
		case "time":
			if len(tmp) > 1 {
				tmp2 := strings.Split(tmp[1], ":")
				if len(tmp2) != 2 {
					http.Error(w, "Error parsing query", 400)
					return
				}

				if tmp2[0] != "instance" {
					http.Error(w, "Unknown parameter : "+tmp2[0], 400)
					return
				}
		
				tmp3 := strings.Split(tmp2[1], "}")
				tmp4 := strings.Split(tmp3[0], "\"")
				instance := tmp4[1]
		
				sendGrafana = queryTime(start, end, step, query, []string{instance})
		
			}else{
				sendGrafana = queryTime(start, end, step, query, listGroup)
			}		
		
		default:
			http.Error(w, "Unknown request : "+tmp[0], 400)
			return
	
	}
	
	sendGrafana.Status = "success"
	
	w.Header().Set("Content-Type", "application/json")
	content, err := json.Marshal(sendGrafana)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Write(content)
}

type Check struct{
	Status string `json:"status"`
	Data []string `json:"data,omitempty"`	
}

func checkAPI(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var c Check
	c.Status = "success"
	c.Data = []string{"counter","time"}
	content, err := json.Marshal(c)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Write(content)

}
