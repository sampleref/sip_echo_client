# sip_example

#### Build binary
go build -o bin/sip_client SipApp.go

#### Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/sip_client_amd64 SipApp.go

#### Build docker image
docker buildx build --platform linux/amd64 --network host -t sip_go_client:1.0 -f ./build/Dockerfile .

