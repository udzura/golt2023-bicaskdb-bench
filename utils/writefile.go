package main

import (
	"log"
	"os"

	"github.com/octu0/bitcaskdb"
)

func main() {
	db, err := bitcaskdb.Open("./db/fileserver")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if len(os.Args) != 3 {
		panic("usage: go writefile.go KEY FILEPATH")
	}

	key := os.Args[1]
	file := os.Args[2]
	r, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	if err := db.Put([]byte(key), r); err != nil {
		panic(err)
	}

	log.Print("OK")
}
