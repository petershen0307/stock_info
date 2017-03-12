package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const urlTemplate string = "https://www.google.com/finance/getprices?q=%[1]s&x=TPE&i=%[2]d&p=%[3]s&f=d,c,h,l,o,v"
const fileNameTemplate string = "%s_%d_%s.json"
const columnFields string = "DATE,CLOSE,HIGH,LOW,OPEN,VOLUME"

type queryInfo struct {
	// stock ID
	stockID string
	/*
		[n]Y: query [n] years
		[n]M: query [n] months
		[n]d: query [n] days
	*/
	duration string
	/*
		every row's interval
		the unit is "Second"
	*/
	interval int64
}

// StockPrice is using by HistoryPrices
// this is the single stock price information
type StockPrice struct {
	Date                   int64
	Close, High, Low, Open float32
	Volume                 int
}

// HistoryPrices is using for JSON structure
type HistoryPrices struct {
	Prices    []StockPrice
	DateIndex []int64
}

func fetchStockPrices(info queryInfo) {
	url := fmt.Sprintf(urlTemplate, info.stockID, info.interval, info.duration)
	fmt.Println(url)
	timeoutRequest := http.Client{Timeout: time.Second * 10}
	response, err := timeoutRequest.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	writeToFile(response.Body, info)
}

func writeToFile(reader io.Reader, info queryInfo) {
	fileName := fmt.Sprintf(fileNameTemplate, info.stockID, info.interval, info.duration)
	fo, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	// Close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	scanner := bufio.NewScanner(reader)
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	isData := false
	var out HistoryPrices
	var dateStart int64
	for scanner.Scan() {
		readStr := scanner.Text()
		// check columns
		if strings.Index(readStr, "COLUMNS") == 0 {
			if 0 != strings.Compare(strings.Split(readStr, "=")[1], columnFields) {
				fmt.Printf("Columns are [%s], not equal with [%s]\n", readStr, columnFields)
				return
			}
			continue
		}

		// line start with "a" is the start date point
		if strings.Index(readStr, "a") == 0 {
			readStr = readStr[1:]
			isData = true
			// get the start date point
			dateStart, err = strconv.ParseInt(strings.Split(readStr, ",")[0], 10, 64)
			if err != nil {
				panic(err)
			}
		}
		if isData {
			var price StockPrice
			fmt.Sscanf(readStr, "%d,%f,%f,%f,%f,%d", &price.Date, &price.Close, &price.High, &price.Low, &price.Open, &price.Volume)
			// calculate date time with start date point
			if dateStart != price.Date {
				price.Date = dateStart + price.Date*info.interval
			}
			out.Prices = append(out.Prices, price)
			out.DateIndex = append(out.DateIndex, price.Date)
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	//fmt.Println(string(b))
	fo.Write(b)
}

func main() {
	fmt.Println("Testing begin")
	fetchStockPrices(queryInfo{
		stockID:  "0050",
		interval: 86400,
		duration: "3d"})
	fmt.Println("Testing end")
}
