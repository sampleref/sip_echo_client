# sip_echo_client
1. Simple Go based sip client to listen for SIP protocol as UAS on port 5063
2. The Media is just echoed, whatever is sent via rtp to UAS is sent back as client source
3. Current supported codec is PCMU and H264
4. Options:  
   **-u**, Default: "local", "SIP Username"    
   **-h**, Default: "0.0.0.0", "sip client host"  
   **-public**, Default: "0.0.0.0", "sip client public host/ip"
   **-p**, Default: 5063, "sip port"   
   **-d**, Default: "info", "Log Level: trace/debug/info"  
   **-mp**, Default: 40100, "Media Port Starts from and next 100 ports" 

#### Build binary
`go build -o bin/sip_client SipApp.go`

#### Build for Linux
`GOOS=linux GOARCH=amd64 go build -o bin/sip_client_amd64 SipApp.go`

#### Build docker image
`docker buildx build --platform linux/amd64 --network host -t sip_go_client:1.0 -f ./build/Dockerfile .`

#### Run binary 
`./sip_client_amd64 -h <LOCAL_IP> -public <PUBLIC_IP> -mp 40100`