build:
	go build -o socketbot main.go

docker: build
	sudo docker build -t socketbot .

start:
	sudo docker run -d --name=socketbot --restart=always socketbot

stop:
	sudo docker rm -f socketbot; true

restart: stop start
	echo "restarted"
