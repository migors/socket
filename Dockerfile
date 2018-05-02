FROM ubuntu:16.04

RUN apt-get update
RUN apt-get install -y ca-certificates

WORKDIR /bot
ADD socketbot socketbot
ADD sockets.kml sockets.kml
ADD token.txt token.txt

CMD ["/bot/socketbot"]
