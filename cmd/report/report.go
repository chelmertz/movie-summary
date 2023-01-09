package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/chelmertz/movie-summary/imdb"
)

// TODO other cmd: web gui for uploading your own .csv (wasm?)
func main() {
	filename := flag.String("im", "", "filename of imdb ratings csv export, only deals with movies")
	flag.Parse()
	if *filename == "" {
		log.Fatalf("Missing filename")
	}

	f, err := os.Open(*filename)
	if err != nil {
		log.Fatalf("File %s is not readable", *filename)
	}

	data, err := imdb.NewFromCsv(f)
	if err != nil {
		log.Fatalf("Could not generate imdb data: %v", err)
	}
	fmt.Printf("success! data:\n\n%+v\n", data)

	// TODO output pretty HTML.. or intermediate json
}
