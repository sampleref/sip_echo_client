package main

import (
	"flag"
	"sip_example/sip_uas"
	"sip_example/sip_util"
)

func main() {
	username := flag.String("u", "local", "SIP Username")
	dialogHost := flag.String("h", "0.0.0.0", "sip client host")
	dialogPort := flag.Int("p", 5063, "sip port")
	logLevel := flag.String("d", "info", "Log Level: trace/debug/info")
	mediaPort := flag.Int("mp", 40100, "Media Port Starts from")
	flag.Parse()
	sip_util.InitLogger(*logLevel)
	sip_util.MediaPortStart = *mediaPort
	sip_util.InitializePorts()
	sip_uas.RunSipUasDialog(*username, *dialogHost, *dialogPort)
}
