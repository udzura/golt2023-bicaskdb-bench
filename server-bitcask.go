package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/octu0/bitcaskdb"
)

var db *bitcaskdb.Bitcask

func main() {
	db_, err := bitcaskdb.Open("./db/fileserver")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db = db_

	http.HandleFunc("/", handler)
	log.Print("info: Server Running: 127.0.0.1:13000")
	http.ListenAndServe(":13000", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path
	key := strings.TrimLeft(filePath, "/")
	value, err := db.Get([]byte(key))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found\n"))
		return
	}
	buf := make([]byte, 0)
	wr := bytes.NewBuffer(buf)
	io.Copy(wr, value)
	rs := bytes.NewReader(wr.Bytes())

	http.ServeContent(w, r, r.URL.Path, time.Now(), rs)
}
