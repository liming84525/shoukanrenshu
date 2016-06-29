package main

import (
	"log"
	"strings"
	"strconv"
	"encoding/json"
	"encoding/csv"
	"os"
	"flag"

	"github.com/go-martini/martini"
	"net/http"
	"time"
	"io"
)

type Count struct {
	Id    string `json:"cid"`
	Count int64 `json:"count"`
}

type Data struct {
	Cid string
	Name string
	Alias string
	Type int
	Weight int
}

var data map[string]Data
var flag_inputFile = flag.String("input", "./std-channel-weight.csv", "flag to read data from")
var flag_outputFile = flag.String("output","./std-channel-weight.csv","flag to create new data")

func main() {

	flag.Parse()

	data, _ = readDataFromFile(*flag_inputFile)

	m := martini.Classic()

	m.Get("/getwatchstatus", func(res http.ResponseWriter, req *http.Request){
		if err := req.ParseForm(); err != nil {
			log.Println("parse fail")
		}
		ret := req.Form.Get("ids")
		ids := strings.Split(ret, ",")
		counts := handle(ids)
		r, err := json.Marshal(counts)
		if err != nil {
			log.Println("parse result to json fail")
		}
		res.Write(r)
	})

	m.RunOnAddr(":8080")

}




//处理逻辑
func handle(ids []string) []Count {
	counts := make([]Count, 0)
	log.Println(ids[0])
	if len(ids) == 1 && strings.EqualFold(ids[0],"") {
		for k , _ := range data {
			//当前时间的浮点数
			pt := getCurve(time.Now())
			c := int64(pt*getWeight(k))
			count := Count{
				Id: k,
				Count: c,
			}
			counts = append(counts, count)
		}
	}
	if len(ids) >= 1 && !strings.EqualFold(ids[0],"") {
		for _, id := range ids {
			//大小写转换
			id = strings.ToUpper(id)
			//验证id合法
			if _, ok := data[id]; !ok {
				continue
			}
			//当前时间的浮点数
			pt := getCurve(time.Now())
			//乘以频道id的权重
			c := int64(pt*getWeight(id))
			count := Count{
				Id:id,
				Count: c,
			}
			counts = append(counts, count)
		}
	}
	return counts
}


//返回曲线的比值 0~1
func getCurve(now time.Time) float64 {
	//last one is a copy of first element, in order to prevent index out of range error
	curve := [25]float64{0.2, 0.15, 0.1, 0.02, 0.01, 0.005, 0.2, 0.3, 0.2, 0.2, 0.1, 0.1, 0.3, 0.3, 0.2, 0.2, 0.3, 0.35, 0.4, 0.7, 0.754, 0.8, 0.6, 0.5, 0.2}
	hour := now.Hour()
	minute := now.Minute()
	interval := curve[hour] + (curve[hour+1]-curve[hour]) * float64(minute) / 60.0
	log.Println(interval)
	return interval
}

//获得id的权重
func getWeight(id string) float64 {
	weight := 0
	if v, ok := data[id]; ok {
		weight = v.Weight
	}
	log.Println(weight)
	return float64(weight);
}

func readDataFromFile(fileName string) (map[string]Data, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModeType)
	log.Println("reading ...", fileName)
	defer file.Close()
	if err != nil {
		log.Println("open file fail")
		data, _ = readDataFromFile("./std-channel-utf8.csv")
		writeFile(*flag_outputFile)
		log.Panic("due to read file fail, create new file, try again")
	}
	reader := csv.NewReader(file)
	reader.Comma = ','
	data := make(map[string]Data)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			log.Println("reading file finish")
			break
		} else if err != nil {
			log.Println("err: ",err)
		}
		item := Data{}
		for idx, col := range line {
			if idx == 0 {
				item.Cid = col
			}
			if idx == 1 {
				item.Alias = col
			}
			if idx == 2 {
				item.Name = col
			}
			if idx == 3 {
				item.Type, _ = strconv.Atoi(col)
			}
			if idx == 4 {
				item.Weight, _ = strconv.Atoi(col)
			}
		}
		data[item.Cid] = item
	}
	return data, nil
}

func writeFile(output string) {
	file, err := os.Create(output)
	defer file.Close()
	if err != nil {
		log.Println("open file fail")
	}
	writer := csv.NewWriter(file)
	for _, item := range data {
		d := make([]string, 0)
		d = append(d, item.Cid, item.Alias, item.Name,strconv.Itoa(item.Type), strconv.Itoa(item.Weight) )
		writer.Write(d)
	}
	writer.Flush()
}


