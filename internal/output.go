package internal

import (
	"github.com/google/gopacket/pcapgo"
	"log"
	"net"
	"net/http"
)

func NewConsoleOutput(capture *Capture) error {
	for packet := range capture.PacketSource.Packets() {
		log.Println(packet)
	}
	return nil
}

func NewTcpServerOutput(capture *Capture, listen string) error {
	addr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		return err
	}
	tcp, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	defer tcp.Close()
	log.Printf("Listening TCP on %s...", listen)
	conn, err := tcp.AcceptTCP()
	if err != nil {
		return err
	}
	defer conn.Close()
	writer, err := pcapgo.NewNgWriter(conn, capture.LinkType)
	if err != nil {
		return err
	}
	defer writer.Flush()
	for packet := range capture.PacketSource.Packets() {
		err := writer.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		if err != nil {
			return err
		}
	}
	return nil
}

func NewHttpServerOutput(capture *Capture, listen string) error {
	connFinished := make(chan bool, 1)
	http.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/x-pcapng")
		writer.Header().Set("Content-Disposition", "attachment; filename=\"capture.pcapng\"")
		writer.WriteHeader(http.StatusOK)
		defer func() { connFinished <- true }()
		ngWriter, err := pcapgo.NewNgWriter(writer, capture.LinkType)
		if err != nil {
			return
		}
		defer ngWriter.Flush()

		for packet := range capture.PacketSource.Packets() {
			err := ngWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
			if err != nil {
				return
			}
		}
	})
	server := &http.Server{
		Addr: listen,
	}
	go func() {
		<-connFinished
		server.Shutdown(nil)
	}()
	log.Printf("Listening HTTP on %s...", listen)
	err := server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
