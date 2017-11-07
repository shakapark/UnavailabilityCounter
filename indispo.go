package main

import(
	"io/ioutil"
	"encoding/json"
	"os"
	"strconv"
	"time"
	
	"github.com/prometheus/common/log"
)

type SaveData struct {
	Month string `json:"month"`
	Year int `json:"year"`
	Data  [][]int64 `json:"data"`    //[[timeStamp1,timeStamp2],[timeStamp1,timeStamp2],[timeStamp1,timeStamp2]]
}

func addData(a [][]int64, n []int64) [][]int64{	
	var t = make([][]int64,len(a)+1)
	for i, v := range a {
		t[i] = v
	}	
	t[len(a)] = n
	return t
}

func newMonth(instance string, groupName string, monthName string, y int){
	_, err := os.Create("/data/"+instance+"/"+groupName+"/"+monthName+strconv.Itoa(y))
	if err != nil {
		log.Fatal("err", err)
	}
}

func newMonthG(instance string, monthName string, y int){
	_, err := os.Create("/data/"+instance+"/"+monthName+strconv.Itoa(y))
	if err != nil {
		log.Fatal("err", err)
	}
}

func register(result int, instance string, groupName string){
	var path string
	indi := Indispos[groupName]
	if result == 0 {
		
		if indi.Progress == false {
			indi.StartTimeStamp = time.Now()
			
			if indi.StartTimeStamp.Month() != indi.TimeStampBack.Month() {
				newMonth(instance, groupName, indi.StartTimeStamp.Month().String(), indi.StartTimeStamp.Year())
			}
			
			indi.Progress = true
			Indispos[groupName] = indi
		}
	
	}else{
		
		if indi.Progress == true {
			indi.StopTimeStamp = time.Now()
			path = "/data/"+instance+"/"+groupName+"/"+indi.StartTimeStamp.Month().String()+strconv.Itoa(indi.StartTimeStamp.Year())
						
			contentFile, err := ioutil.ReadFile(path)
			if err != nil {
				log.Infoln("err : ", err)
				_, err = os.Create(path)
				if err != nil {
					log.Fatal("err", err)
				}
				log.Infoln("msg : ", "File create : ", path)
			}

			var saveData SaveData

			if contentFile == nil || len(contentFile) == 0 {
				saveData.Month = indi.StartTimeStamp.Month().String()
				saveData.Year = indi.StartTimeStamp.Year()
				saveData.Data = [][]int64{{indi.StartTimeStamp.Unix(), indi.StopTimeStamp.Unix()}}
			}else{
				if err := json.Unmarshal(contentFile, &saveData); err != nil {
					log.Fatal("err", err)
				}
				saveData.Data = addData(saveData.Data, []int64{indi.StartTimeStamp.Unix(),indi.StopTimeStamp.Unix()})
			}
			
			contentFile , err = json.Marshal(saveData)
			if err != nil {
				log.Fatal(err)
			}
			
			f, err := os.OpenFile(path, os.O_RDWR, 0755)
			if err != nil {
				log.Fatal(err)
			}
			
			_, err = f.Write(contentFile)
			if err != nil {
				log.Fatal(err)
			}
			
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
			
			indi.Progress = false
			indi.TimeStampBack = indi.StartTimeStamp
			
			Indispos[groupName] = indi
		}
	}
}

func registerG(result bool, instance string){
	var path string
	indi := Indispos[instance]
	if result {
		
		if indi.Progress == false {
			indi.StartTimeStamp = time.Now()
			
			if indi.StartTimeStamp.Month() != indi.TimeStampBack.Month() {
				newMonthG(instance, indi.StartTimeStamp.Month().String(), indi.StartTimeStamp.Year())
			}
			
			indi.Progress = true
			Indispos[instance] = indi
		}
	
	}else{
		
		if indi.Progress == true {
			indi.StopTimeStamp = time.Now()
			path = "/data/"+instance+"/"+indi.StartTimeStamp.Month().String()+strconv.Itoa(indi.StartTimeStamp.Year())
						
			contentFile, err := ioutil.ReadFile(path)
			if err != nil {
				log.Infoln("err : ", err)
				_, err = os.Create(path)
				if err != nil {
					log.Fatal("err", err)
				}
				log.Infoln("msg : ", "File create : ", path)
			}

			var saveData SaveData

			if contentFile == nil || len(contentFile) == 0 {
				saveData.Month = indi.StartTimeStamp.Month().String()
				saveData.Year = indi.StartTimeStamp.Year()
				saveData.Data = [][]int64{{indi.StartTimeStamp.Unix(), indi.StopTimeStamp.Unix()}}
			}else{
				if err := json.Unmarshal(contentFile, &saveData); err != nil {
					log.Fatal("err", err)
				}
				saveData.Data = addData(saveData.Data, []int64{indi.StartTimeStamp.Unix(),indi.StopTimeStamp.Unix()})
			}
			
			contentFile , err = json.Marshal(saveData)
			if err != nil {
				log.Fatal(err)
			}
			
			f, err := os.OpenFile(path, os.O_RDWR, 0755)
			if err != nil {
				log.Fatal(err)
			}
			
			_, err = f.Write(contentFile)
			if err != nil {
				log.Fatal(err)
			}
			
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
			
			indi.Progress = false
			indi.TimeStampBack = indi.StartTimeStamp
			
			Indispos[instance] = indi
		}
	}
}
