build:
	go build -o bin/socketbot bitbucket.org/pav5000/socketbot/cmd/bot

docker: build
	sudo docker build -t socketbot .

run: docker
	sudo docker run \
		--rm -ti \
		-v `pwd`/data:/bot/data  \
		socketbot

start:
	sudo docker run -d --name=socketbot --restart=always -v `pwd`/data:/bot/data socketbot

stop:
	sudo docker rm -f socketbot; true

restart: stop start
	echo "restarted"

deploy:
	ssh linode -t 'bash -l -c "cd ~/docker/socketbot && git pull && make docker && make restart"'

.PHONY: sockets
sockets:
	rsync --progress data/www/*.* pavl.uk:~/docker/socketbot/data/www
