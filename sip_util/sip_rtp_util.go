package sip_util

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
)

var (
	MediaPortStart = 40100
	MaxPortsNeeded = 100
	MediaPorts     sync.Map
)

func InitializePorts() {
	for index := 0; index < MaxPortsNeeded; index++ {
		MediaPort := MediaPortStart + index
		MediaPorts.Store(MediaPort, true)
	}
}

func FindNextFreePort() int {
	FreePortInRange := 0
	for {
		FreePortInRange = findNextRangePort()
		if FreePortInRange == 0 {
			log.Error().Msg("No free port found")
			break
		}
		conn, err := net.ListenUDP("udp", &net.UDPAddr{
			Port: FreePortInRange,
			IP:   net.ParseIP("0.0.0.0"),
		})
		if err != nil {
			log.Err(err)
			log.Error().Msg("Port " + fmt.Sprint(FreePortInRange) + " Looks occupied, trying other")
			conn.Close()
		} else {
			conn.Close()
			log.Info().Msg("Port " + fmt.Sprint(FreePortInRange) + " Allocated")
			return FreePortInRange
		}
	}
	return FreePortInRange
}

func findNextRangePort() int {
	FreePortInRange := 0
	MediaPorts.Range(func(key, value interface{}) bool {
		if value.(bool) {
			FreePortInRange = key.(int)
			return false
		}
		return true
	})
	if FreePortInRange > 0 {
		log.Info().Msg("Port " + fmt.Sprint(FreePortInRange) + " Occupied")
		MediaPorts.Store(FreePortInRange, false)
	}
	return FreePortInRange
}

func ReleasePort(port int) {
	log.Info().Msg("Port " + fmt.Sprint(port) + " Released")
	_, ok := MediaPorts.Load(port)
	if ok {
		MediaPorts.Store(port, true)
	}
}
