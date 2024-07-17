package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

var (
	filename        = flag.String("file", "", "filename")
	shouldWriteBack = flag.Bool("w", false, "write back")
	version         = flag.String("version", "", "print version")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("sortimport:")

	if strings.Contains(strings.Join(os.Args, " "), "-version") {
		log.Println(_version)
		os.Exit(0)
	}

	flag.Parse()

	if *filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	handler := newHandler(*filename)
	handler.start()
	if *shouldWriteBack {
		handler.writeBack()
	} else {
		handler.print()
	}

}
