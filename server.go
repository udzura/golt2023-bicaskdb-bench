package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/octu0/bitcaskdb"
)

var db *bitcaskdb.Bitcask

func main() {
	db_, err := bitcaskdb.Open("./db/fileserver", bitcaskdb.WithMaxDatafileSize(1024*1024*1024))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db = db_

	http.HandleFunc("/files/", handler)
	http.HandleFunc("/bitcask/", handlerBC)
	log.Print("info: Server Running: 127.0.0.1:13000")
	http.ListenAndServe(":13000", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "get", "GET":
		onGet(w, r)
	case "post", "POST":
		onPost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Not Supported\n"))
	}
}

func onGet(w http.ResponseWriter, r *http.Request) {
	filePath := "./data" + strings.TrimPrefix(r.URL.Path, "/files")
	file, err := os.Open(filePath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found\n"))
		return
	}
	defer file.Close()

	http.ServeContent(w, r, r.URL.Path, time.Now(), file)
}

func onPost(w http.ResponseWriter, r *http.Request) {
	filePath := "./data" + strings.TrimPrefix(r.URL.Path, "/files")
	file, err := os.Create(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer file.Close()
	length, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if _, err := io.CopyN(file, r.Body, length); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func handlerBC(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "get", "GET":
		onGetBC(w, r)
	case "post", "POST":
		onPostBC(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Not Supported\n"))
	}
}

func onGetBC(w http.ResponseWriter, r *http.Request) {
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

func onPostBC(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path
	key := strings.TrimLeft(filePath, "/")
	length, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	buf := bytes.NewBuffer([]byte{})
	if _, err := io.CopyN(buf, r.Body, length); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer r.Body.Close()

	if err := db.PutBytes([]byte(key), buf.Bytes()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
