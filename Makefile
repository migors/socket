build:
	sudo docker build -t socketbot .

run: build
	sudo docker run --rm -ti -v `pwd`/data:/bot/data socketbot

start:
	sudo docker run -d --name=socketbot --restart=always -v `pwd`/data:/bot/data socketbot

stop:
	sudo docker rm -f socketbot; true

restart: stop start
	echo "restarted"
