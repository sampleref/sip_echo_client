package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"sip_echo_client/sip_uas"
	"sip_echo_client/sip_util"
	"strings"
)

func main() {
	username := flag.String("u", "local", "SIP Username")
	dialogHost := flag.String("h", "0.0.0.0", "sip client host")
	publicHost := flag.String("public", *dialogHost, "sip client public host/ip")
	dialogPort := flag.Int("p", 5063, "sip port")
	logLevel := flag.String("d", "info", "Log Level: trace/debug/info")
	mediaPort := flag.Int("mp", 40100, "Media Port Starts from")
	flag.Parse()
	sip_util.InitLogger(*logLevel)
	sip_util.MediaPortStart = *mediaPort
	sip_util.InitializePorts()
	sip_util.StartCounters()
	sip_uas.StartSipServer(*username, *dialogHost, *dialogPort, *publicHost)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("->>>>>>> q:Quit >>>>>> r:RTP Stats All >>>>>> rl:RTP Stats Live >>>>>: \n")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("q", text) == 0 {
			log.Info().Msg("Shutting down...")
			sip_uas.StopSipServer()
			sip_util.StopRTPStats()
			log.Info().Msg("Shutting Complete!")
			os.Exit(0)
		}
		if strings.Compare("r", text) == 0 {
			sip_util.PrintAllStats = !sip_util.PrintAllStats
		}
		if strings.Compare("rl", text) == 0 {
			sip_util.PrintLiveStats = !sip_util.PrintLiveStats
		}
	}
}
