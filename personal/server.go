package main

import (
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe("jprbnsn.com", http.FileServer(http.Dir("/home/jp/www"))))
}