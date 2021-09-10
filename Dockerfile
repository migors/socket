FROM ubuntu:20.04

RUN apt update && apt install -y ca-certificates

WORKDIR /bot
ADD bin/socketbot /bot/socketbot
# ENV HTTP_PROXY 192.168.2.1:3128
# ENV HTTPS_PROXY 192.168.2.1:3128

CMD ["/bot/socketbot"]
