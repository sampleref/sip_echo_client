package sip_uas

import (
	"errors"
	"github.com/rs/zerolog/log"
	"net"
	"sip_echo_client/sip_util"
	"strconv"
)

type RTPSession struct {
	callId string

	rVideoHost string
	rAudioHost string
	rVideoPort int
	rAudioPort int

	lVideoHost string
	lAudioHost string
	lVideoPort int
	lAudioPort int

	keepRunning bool
	reInvite    bool
}

func (rtpSession *RTPSession) StopRTPListeners() {
	rtpSession.keepRunning = false
	sip_util.ReleasePort(rtpSession.lAudioPort)
	sip_util.ReleasePort(rtpSession.lVideoPort)
}

func (rtpSession *RTPSession) StartRTPListenerAudioVideo() (int, int, error) {
	rtpSession.keepRunning = true

	//Audio Listener
	rtpSession.lAudioPort = sip_util.FindNextFreePort()
	if rtpSession.lAudioPort == 0 {
		log.Error().Msg("Audio: Failed to find a free port")
		return 0, 0, errors.New("audio: failed to find a free port")
	}
	aConn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: rtpSession.lAudioPort,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		log.Err(err)
		return 0, 0, err
	}

	go func() {
		buff := make([]byte, 1500)

		for rtpSession.keepRunning {
			n, raddr, err := aConn.ReadFromUDP(buff)
			if err != nil {
				log.Err(err)
				return
			}
			sip_util.AudioPckRecv <- rtpSession.callId
			aConn.WriteToUDP(buff[:n], raddr)
			sip_util.AudioPckSent <- rtpSession.callId
		}
		log.Info().Msg("Closing audio udp for callId " + rtpSession.callId)
		aConn.Close()
	}()

	_, ok := aConn.LocalAddr().(*net.UDPAddr)
	if !ok {
		log.Error().Msg("Audio RTP: Failed to cast *net.UDPAddr")
		return 0, 0, errors.New("failed to cast *net.UDPAddr")
	}

	//Video Listener
	rtpSession.lVideoPort = sip_util.FindNextFreePort()
	if rtpSession.lVideoPort == 0 {
		log.Error().Msg("Video: Failed to find a free port")
		return 0, 0, errors.New("video: failed to find a free port")
	}
	vConn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: rtpSession.lVideoPort,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		log.Err(err)
		return 0, 0, err
	}

	go func() {
		buff := make([]byte, 1500)

		for rtpSession.keepRunning {
			n, raddr, err := vConn.ReadFromUDP(buff)
			if err != nil {
				log.Err(err)
				return
			}
			sip_util.VideoPckRecv <- rtpSession.callId
			vConn.WriteToUDP(buff[:n], raddr)
			sip_util.VideoPckSent <- rtpSession.callId
		}
		log.Info().Msg("Closing video udp for callId " + rtpSession.callId)
		vConn.Close()
	}()

	_, ok = vConn.LocalAddr().(*net.UDPAddr)
	if !ok {
		log.Error().Msg("Video RTP: Failed to cast *net.UDPAddr")
		return 0, 0, errors.New("failed to cast *net.UDPAddr")
	}

	log.Info().Msg("Selected CallId " + rtpSession.callId + " Audio Port: " + strconv.Itoa(rtpSession.lAudioPort) +
		" Video: " + strconv.Itoa(rtpSession.lVideoPort))
	return rtpSession.lAudioPort, rtpSession.lVideoPort, nil
}
