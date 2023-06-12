package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/montanaflynn/stats"
	"github.com/octu0/bitcaskdb"
	"github.com/olekukonko/tablewriter"
)

var data []byte

const N = 500

func init() {
	data = make([]byte, 5*1024)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
}

func main() {
	elapCh := make(chan time.Duration)
	resData := make([]float64, 0)
	ctx, stop := context.WithCancel(context.Background())

	var err error = nil
	go func() {
		if os.Args[1] == "files" {
			err = benchFiles(elapCh)
		} else {
			err = benchBitcaskdb(elapCh)
		}

		if err != nil {
			panic(err)
		}
		stop()
	}()

	loop := true
	for loop {
		select {
		case r := <-elapCh:
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
	p995, _ := stats.Percentile(resData, 99.5)
	max, _ := stats.Max(resData)
	ms := float64(time.Millisecond)

	fmt.Println("")

	fmt.Println("")
	buf := bytes.NewBufferString("")
	tbl := tablewriter.NewWriter(buf)
	tbl.SetHeader([]string{"AVG", "50%ILE", "90%ILE", "95%ILE", "99.5%ILE", "MAX", "STDDEV"})
	tbl.SetAutoWrapText(false)
	tbl.SetBorders(tablewriter.Border{Left: false, Right: false, Top: false, Bottom: false})
	tbl.SetRowLine(true)
	rec := make([]string, 0, 8)
	rec = append(rec, fmt.Sprintf("%.3fms", mean/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", p50/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", p90/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", p95/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", p995/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", max/ms))
	rec = append(rec, fmt.Sprintf("%.3fms", stddev/ms))
	tbl.Append(rec)

	tbl.Render()
	fmt.Printf("%s\n", buf.String())
}

func benchFiles(ch chan time.Duration) error {
	for i := 0; i < N; i++ {
		dirname := uuid.New().String()
		os.MkdirAll("./data/"+dirname, 0o0755)
		f, err := os.Create(filepath.Join("./data", dirname, "dummy.image"))
		if err != nil {
			return err
		}
		start := time.Now()
		if _, err := f.Write(data); err != nil {
			return err
		}
		ch <- time.Since(start)
		f.Sync()
		f.Close()
	}
	return nil
}

func benchBitcaskdb(ch chan time.Duration) error {
	db, err := bitcaskdb.Open("./db/", bitcaskdb.WithMaxDatafileSize(1024*1024*1024))
	if err != nil {
		return err
	}
	defer db.Close()

	for i := 0; i < N; i++ {
		dirname := uuid.New().String()
		key := filepath.Join("data", dirname, "dummy.image")
		start := time.Now()

		if err := db.PutBytes([]byte(key), data); err != nil {
			return err
		}

		ch <- time.Since(start)
	}
	return nil
}
