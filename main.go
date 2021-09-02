package main

import (
	"flag"
	"fmt"
	"github.com/philipglazman/sprinter/sprinter"
	"log"
)

func main() {
	// Use a cli flag to start the crawler.
	var root string
	flag.StringVar(&root, "root", "", "the root domain to crawl")

	flag.Parse()

	s, err := sprinter.NewSprinter(root)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("starting sprinter with root %s\n", s.Root())

	res := s.Crawl()

	fmt.Printf("%s\n", res)
}
