package sip_util

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"
	"os"
	"sync/atomic"
	"time"
)

var (
	VideoPckRecv = make(chan string, 1000)
	VideoPckSent = make(chan string, 1000)
	AudioPckRecv = make(chan string, 1000)
	AudioPckSent = make(chan string, 1000)

	CallStatus      = make(map[string]*CallStats)
	PrintLiveStats  = false
	PrintAllStats   = false
	runStatsPrinter = true
)

const (
	CALL_STATE_STARTED = "STARTED"
	CALL_STATE_STOPPED = "STOPPED"
)

type CallStats struct {
	callId      string
	vidRecvPkts atomic.Uint64
	vidSentPkts atomic.Uint64
	audRecvPkts atomic.Uint64
	audSentPkts atomic.Uint64
	state       string
}

func NewCallStats(callId string) {
	CallStatus[callId] = &CallStats{callId: callId, state: CALL_STATE_STARTED}
}

func StopCallStatus(callId string) {
	stats := CallStatus[callId]
	if stats != nil {
		stats.state = CALL_STATE_STOPPED
	}
}

func StartCounters() {
	go func() {
		for callId := range VideoPckRecv {
			stats := CallStatus[callId]
			if stats != nil {
				stats.vidRecvPkts.Add(1)
			}
		}
	}()
	go func() {
		for callId := range VideoPckSent {
			stats := CallStatus[callId]
			if stats != nil {
				stats.vidSentPkts.Add(1)
			}
		}
	}()
	go func() {
		for callId := range AudioPckSent {
			stats := CallStatus[callId]
			if stats != nil {
				stats.audSentPkts.Add(1)
			}
		}
	}()
	go func() {
		for callId := range AudioPckRecv {
			stats := CallStatus[callId]
			if stats != nil {
				stats.audRecvPkts.Add(1)
			}
		}
	}()
	PrintRTPStats()
	log.Info().Msg("Started RTP packet counters...")
}

func StopRTPStats() {
	close(VideoPckRecv)
	close(VideoPckSent)
	close(AudioPckRecv)
	close(AudioPckSent)
	runStatsPrinter = false
	log.Info().Msg("Stopped RTP packet counters...")
}

func PrintRTPStats() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredBright)
	t.AppendHeader(table.Row{"Call Id", "Audio Tx", "Audio Rx", "Video Tx", "Video Rx", "State"})
	go func() {
		for runStatsPrinter {
			for callId := range CallStatus {
				CallStats := CallStatus[callId]
				if PrintLiveStats || PrintAllStats {
					if PrintLiveStats {
						if CallStats.state != CALL_STATE_STOPPED {
							t.AppendRow([]interface{}{CallStats.callId, CallStats.audSentPkts, CallStats.audRecvPkts,
								CallStats.vidSentPkts, CallStats.vidRecvPkts, CallStats.state})
							t.AppendSeparator()
						}
					} else {
						t.AppendRow([]interface{}{CallStats.callId, CallStats.audSentPkts, CallStats.audRecvPkts,
							CallStats.vidSentPkts, CallStats.vidRecvPkts, CallStats.state})
						t.AppendSeparator()
					}
				}
			}
			if PrintLiveStats || PrintAllStats {
				t.Render()
				log.Info().Msg("")
			}
			t.ResetRows()
			time.Sleep(5 * time.Second)
		}
	}()
}
