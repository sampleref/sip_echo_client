package sip_util

import (
	"testing"
	"time"
)

func TestRTPStatsPrinter(t *testing.T) {

	InitLogger("INFO")
	PrintLiveStats = true
	NewCallStats("CallId001")
	NewCallStats("CallId002")
	NewCallStats("CallId003")
	StartCounters()
	StopCallStatus("CallId003")

	for i := 0; i < 1000; i++ {
		VideoPckSent <- "CallId001"
		VideoPckSent <- "CallId002"
		VideoPckRecv <- "CallId001"
		VideoPckRecv <- "CallId002"

		AudioPckSent <- "CallId001"
		AudioPckSent <- "CallId002"
		AudioPckRecv <- "CallId001"
		AudioPckRecv <- "CallId002"
	}

	PrintLiveStats = false
	PrintAllStats = true
	for i := 0; i < 1000; i++ {
		VideoPckSent <- "CallId001"
		VideoPckSent <- "CallId002"
		VideoPckRecv <- "CallId001"
		VideoPckRecv <- "CallId002"

		AudioPckSent <- "CallId001"
		AudioPckSent <- "CallId002"
		AudioPckRecv <- "CallId001"
		AudioPckRecv <- "CallId002"
	}

	time.Sleep(10 * time.Second)
	StopRTPStats()
}
