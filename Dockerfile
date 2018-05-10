FROM golang:1.10.2 as builder
WORKDIR /go/src/github.com/pav5000/socketbot/
RUN go get github.com/golang/geo/s2 && \
	go get github.com/mattn/go-sqlite3 && \
	go get github.com/jmoiron/sqlx
COPY . /go/src/github.com/pav5000/socketbot/
RUN GOPATH=/go GOOS=linux go build -o socketbot main.go


FROM ubuntu:16.04

RUN apt-get update
RUN apt-get install -y ca-certificates

WORKDIR /bot
COPY --from=builder /go/src/github.com/pav5000/socketbot/socketbot .
ADD token.txt token.txt
# ENV HTTP_PROXY 192.168.2.1:3128
# ENV HTTPS_PROXY 192.168.2.1:3128

CMD ["/bot/socketbot"]
