package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", handler)
	log.Print("info: Server Running: 127.0.0.1:13000")
	http.ListenAndServe(":13000", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	filePath := "./data" + r.URL.Path
	file, err := os.Open(filePath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found\n"))
		return
	}
	defer file.Close()

	http.ServeContent(w, r, r.URL.Path, time.Now(), file)
}
