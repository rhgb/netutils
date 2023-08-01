package internal

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"runtime"
)

type Capture struct {
	dataSource   gopacket.PacketDataSource
	LinkType     layers.LinkType
	PacketSource *gopacket.PacketSource
}

func StartCapturePcap(ifName string, bpfFilter string) (*Capture, error) {
	ifDevName := ""
	if runtime.GOOS == "windows" {
		// Windows uses the interface description, not the name
		ifs, err := pcap.FindAllDevs()
		if err != nil {
			return nil, err
		}
		for _, ifi := range ifs {
			if ifi.Description == ifName {
				ifDevName = ifi.Name
				break
			}
		}
		if ifDevName == "" {
			return nil, errors.New("interface not found")
		}
	} else {
		ifDevName = ifName
	}
	if handle, err := pcap.OpenLive(ifDevName, 2000, true, pcap.BlockForever); err != nil {
		return nil, err
	} else {
		if bpfFilter != "" {
			if err := handle.SetBPFFilter(bpfFilter); err != nil {
				return nil, err
			}
		}
		return &Capture{
			dataSource:   handle,
			LinkType:     handle.LinkType(),
			PacketSource: gopacket.NewPacketSource(handle, handle.LinkType()),
		}, nil
	}
}

func (c *Capture) Close() {
	if handle, ok := c.dataSource.(*pcap.Handle); ok {
		handle.Close()
	}
}
