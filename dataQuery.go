package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const urlTemplate string = "https://www.google.com/finance/getprices?q=%[1]s&x=TPE&i=%[2]d&p=%[3]s&f=d,c,h,l,o,v"
const fileNameTemplate string = "%s_%d_%s.csv"

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
	interval uint
}

func fetchStockInfo(info queryInfo) {
	url := fmt.Sprintf(urlTemplate, info.stockID, info.interval, info.duration)
	fmt.Println(url)
	timeoutRequest := http.Client{Timeout: time.Second * 10}
	response, err := timeoutRequest.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	writeToFile(response.Body, fmt.Sprintf(fileNameTemplate, info.stockID, info.interval, info.duration))
}

func writeToFile(reader io.Reader, fileName string) {
	fo, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
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
	for scanner.Scan() {
		readStr := scanner.Text() + "\n"
		// write table header
		if strings.Index(readStr, "COLUMNS") == 0 {
			fo.WriteString(strings.Split(readStr, "=")[1])
		}

		if isData {
			fo.WriteString(readStr)
		} else if strings.Index(readStr, "a") == 0 {
			isData = true
			fo.WriteString(readStr[1:])
		}
	}
}

func main() {
	fmt.Println("Testing begin")
	fetchStockInfo(queryInfo{
		stockID:  "0050",
		interval: 86400,
		duration: "1M"})
	fmt.Println("Testing end")
}
