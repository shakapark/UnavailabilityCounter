package data

//Data Represent the data in a data file
type Data struct {
	Month string    `json:"month"`
	Year  int       `json:"year"`
	Data  [][]int64 `json:"data"` //[[timeStamp1,timeStamp2],[timeStamp1,timeStamp2],[timeStamp1,timeStamp2]]
}
