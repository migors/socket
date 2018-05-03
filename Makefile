build: deps
	go build -o socketbot main.go

deps:
	go get github.com/golang/geo/s2

run: build
	./socketbot

docker: build
	sudo docker build -t socketbot .

start:
	sudo docker run -d --name=socketbot --restart=always socketbot

stop:
	sudo docker rm -f socketbot; true

restart: stop start
	echo "restarted"
