package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

var (
	filename    = flag.String("file", "", "filename")
	writeback   = flag.Bool("writeback", false, "writeback to the file")
	stdOut      = flag.Bool("std-out", true, "print info into stdout")
	onlyChanged = flag.Bool("only-changed", false, "just print changed line, false will print all info")
	version     = flag.String("version", "", "print version")
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
	if *writeback || *stdOut {
		handler.writeBack()
	} else {
		handler.print()
	}

}
