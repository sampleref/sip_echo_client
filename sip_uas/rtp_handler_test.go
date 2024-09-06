package sip_uas

import (
	"github.com/rs/zerolog/log"
	"net"
	"testing"
)

func TestUDPListenPortsConflict(t *testing.T) {
	testPort := 41000
	msgChan := make(chan string)
	var conn1, conn2 net.PacketConn
	var err error

	go func(name string) {
		conn1, err = net.ListenUDP("udp", &net.UDPAddr{
			Port: testPort,
			IP:   net.ParseIP("0.0.0.0"),
		})
		if err != nil {
			log.Err(err)
			msgChan <- err.Error() + " from " + name
		}
	}("port1")

	go func(name string) {
		conn2, err = net.ListenUDP("udp", &net.UDPAddr{
			Port: testPort,
			IP:   net.ParseIP("0.0.0.0"),
		})
		if err != nil {
			log.Err(err)
			if err != nil {
				log.Err(err)
				msgChan <- err.Error() + " from " + name
			}
		}
	}("port2")

	select {
	case msg := <-msgChan:
		log.Info().Msg("Response from " + msg)
		return
	}
	conn1.Close()
	conn2.Close()
}
