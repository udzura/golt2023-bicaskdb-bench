package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	//go:embed testdata/*.image
	testdata embed.FS
)

func main() {
	subdir := os.Args[1]

	start := time.Now()
	for j := 0; j < 100; j++ {
		for i := 1; i <= 100; i++ {
			name := fmt.Sprintf("data-%03d.image", i)
			f, err := testdata.Open("testdata/" + name)
			if err != nil {
				panic(err)
			}
			s, err := f.Stat()
			if err != nil {
				panic(err)
			}
			size := s.Size()

			url := "http://127.0.0.1:13000/" + subdir + "/" + name
			req, err := http.NewRequest("POST", url, f)
			req.ContentLength = size

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			//log.Println(res.Status)
			//io.Copy(os.Stderr, res.Body)
			res.Body.Close()
		}
	}

	log.Println("elapsed: ", time.Since(start))
}
