FROM ubuntu:20.04

RUN apt update && apt install -y ca-certificates

WORKDIR /bot
ADD bin/socketbot /bot/socketbot

CMD ["/bot/socketbot"]
