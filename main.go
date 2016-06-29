package main

import (
	"log"
	"strings"
	"strconv"
	"encoding/json"
	"encoding/csv"
	"math/rand"
	"math"
	"os"

	"github.com/go-martini/martini"
	"net/http"
	"time"
	"io"
)

type Counts struct {
	Id    string `json:"cid,omitempty"`
	Count int64 `json:"count,omitempty"`
}

type Data struct {
	Cid string
	Name string
	Alias string
	Type int
	Weight int
}

var data []Data

func main() {

	data = openfile()

	m := martini.Classic()

	m.Get("/getwatchstatus", func(res http.ResponseWriter, req *http.Request){
		if err := req.ParseForm(); err != nil {
			log.Println("parse fail")
		}
		ret := req.Form.Get("ids")
		ids := strings.Split(ret, ",")
		counts := handle(ids)
		//writefile()
		r, err := json.Marshal(counts)
		if err != nil {
			log.Println("parse result to json fail")
		}
		res.Write(r)
	})

	m.RunOnAddr(":8080")

}




//处理逻辑
func handle(ids []string) []Counts {

	counts := make([]Counts, 0)
	for _, id := range ids {
		//当前时间的浮点数
		pt := getCurve(time.Now())
		//乘以频道id的权重
		c := int64(pt*getWeight(id))
		count := Counts{
			Id:id,
			Count: c,
		}
		counts = append(counts, count)
	}
	return counts
}


//返回曲线的比值 0~1
func getCurve(now time.Time) float64 {
	curve := [24]float64{0.2, 0.15, 0.1, 0.02, 0.01, 0.005, 0.2, 0.3, 0.2, 0.2, 0.1, 0.1, 0.3, 0.3, 0.2, 0.2, 0.3, 0.35, 0.4, 0.7, 0.754, 0.8, 0.6, 0.5}
	hour := now.Hour()
	minute := now.Minute()
	interval := 0.0
	if curve[hour] < curve[hour+1] {
		interval = curve[hour] + math.Abs(curve[hour+1]-curve[hour]) * float64(minute) / 60.0
	} else if curve[hour] > curve[hour+1] {
		interval = curve[hour] - math.Abs(curve[hour+1]-curve[hour]) * float64(minute) / 60.0
	} else if curve[hour] == curve[hour+1] {
		interval = curve[hour]
	}
	log.Println(interval)
	return interval
}

//获得id的权重
func getWeight(id string) float64 {
	tp := 0
	weight := 0
	for _, item := range data {
		if item.Cid == id {
			tp = item.Type
			weight = item.Weight
			break
		}
	}
	if weight == 0 {
		weight = 99
		switch tp {
		case 1:
			weight *= 100
		case 2:
			weight *= 32
		case 3:
			weight *= 15
		case 4:
			weight *= 5
		default:
			weight *= rand.Intn(5)
		}
	}
	log.Println(weight)
	return float64(weight);
}

func openfile() []Data {
	file, err := os.OpenFile("./std-channel-weight.csv", os.O_RDWR, os.ModeType)
	defer file.Close()
	if err != nil {
		log.Println("open file fail")
	}
	reader := csv.NewReader(file)
	reader.Comma = ','
	data := make([]Data, 0)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			log.Println("reading file finish")
			break
		} else if err != nil {
			log.Println("err: ",err)
		}
		item := Data{}
		for idx, row := range line {
			if idx == 0 {
				item.Cid = row
			}
			if idx == 1 {
				item.Alias = row
			}
			if idx == 2 {
				item.Name = row
			}
			if idx == 3 {
				item.Type, _ = strconv.Atoi(row)
			}
			if idx == 4 {
				item.Weight, _ = strconv.Atoi(row)
			}
		}
		data = append(data, item)
	}
	return data
}

func writefile() {
	file, err := os.Create("std-channel-weight.csv")
	defer file.Close()
	if err != nil {
		log.Println("open file fail")
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	for _, item := range data {
		d := make([]string, 0)
		d = append(d, item.Cid, item.Alias, item.Name,strconv.Itoa(item.Type), strconv.Itoa(item.Weight) )
		writer.Write(d)
	}

}


