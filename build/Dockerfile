# used for build local spec client
FROM --platform=linux/amd64 public.ecr.aws/docker/library/node:18-bookworm
RUN apt update && apt install -y nano net-tools nmap iputils-ping tcpdump sngrep

RUN mkdir -p /home/sip/client
ADD ./bin/sip_client_amd64 /home/sip/client
WORKDIR /home/sip/client

CMD ["/home/sip/client/sip_client_amd64"]