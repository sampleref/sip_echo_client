package main

import (
	"flag"
	"sip_echo_client/sip_uas"
	"sip_echo_client/sip_util"
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
	sip_uas.RunSipUasDialog(*username, *dialogHost, *dialogPort, *publicHost)
}
