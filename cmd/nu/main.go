package main

import (
	"flag"
	"github.com/rhgb/netutils/internal"
	"log"
	"os"
	"os/signal"
)

func main() {
	flagIf := flag.String("i", "", "Interface name to capture on")
	flagOutput := flag.String("o", "console", "Output type, available: console")
	flagListen := flag.String("l", "", "Listen address for output type 'http_server' and 'tcp_server'")
	flagFilter := flag.String("f", "", "BPF filter clause")
	flag.Parse()

	if *flagIf == "" {
		flag.PrintDefaults()
		return
	}
	capture, err := internal.StartCapturePcap(*flagIf, *flagFilter)
	defer capture.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("Capturing interrupted by %s, exiting...", sig)
			capture.Close()
		}
	}()

	if err != nil {
		log.Fatal(err)
	}

	switch *flagOutput {
	case "console":
		err = internal.NewConsoleOutput(capture)
	case "tcp_server":
		err = internal.NewTcpServerOutput(capture, *flagListen)
	case "http_server":
		err = internal.NewHttpServerOutput(capture, *flagListen)
	default:
		flag.PrintDefaults()
	}
	if err != nil {
		log.Fatal(err)
	}
}
