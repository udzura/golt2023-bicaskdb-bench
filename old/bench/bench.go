package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/olekukonko/tablewriter"
)

var (
	//go:embed testdata/*.image
	testdata embed.FS
)

func main() {
	subdir := os.Args[1]
	ech := make(chan time.Duration)
	resData := make([]float64, 0)
	ctx, stop := context.WithCancel(context.Background())

	go func() {
		for j := 0; j < 10; j++ {
			for i := 1; i <= 100; i++ {
				name := fmt.Sprintf("data-%03d.image", i)
				f, err := testdata.Open("testdata/data.image")
				if err != nil {
					panic(err)
				}
				s, err := f.Stat()
				if err != nil {
					panic(err)
				}
				size := s.Size()

				url := "http://127.0.0.1:13000/" + subdir + "/" + name
				start := time.Now()
				req, err := http.NewRequest("POST", url, f)
				req.ContentLength = size

				client := &http.Client{}
				res, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				if res.StatusCode != 200 {
					log.Println(res.Status)
				}
				//io.Copy(os.Stderr, res.Body)
				res.Body.Close()
				go func(e time.Duration) {
					ech <- e
				}(time.Since(start))
			}
		}
		stop()
	}()

	loop := true
	for loop {
		select {
		case r := <-ech:
			resData = append(resData, float64(r))
		case <-ctx.Done():
			loop = false
		}
	}

	mean, _ := stats.Mean(resData)
	stddev, _ := stats.StandardDeviation(resData)
	p50, _ := stats.Percentile(resData, 50)
	p90, _ := stats.Percentile(resData, 90)
	p95, _ := stats.Percentile(resData, 95)
	p99, _ := stats.Percentile(resData, 95)
	p995, _ := stats.Percentile(resData, 99.5)
	max, _ := stats.Max(resData)
	ms := float64(time.Millisecond)

	fmt.Println("")

	fmt.Println("")
	buf := bytes.NewBufferString("")
	tbl := tablewriter.NewWriter(buf)
	tbl.SetHeader([]string{"AVG", "50%ILE", "90%ILE", "95%ILE", "99%ILE", "99.5%ILE", "MAX", "STDDEV"})
	tbl.SetAutoWrapText(false)
	tbl.SetBorders(tablewriter.Border{false, false, false, false})
	tbl.SetRowLine(true)
	rec := make([]string, 0, 8)
	rec = append(rec, fmt.Sprintf("%.3fms", mean/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", p50/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", p90/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", p95/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", p99/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", p995/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", max/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", stddev/ms))
	tbl.Append(rec)

	tbl.Render()
	fmt.Printf("%s\n", buf.String())
}
