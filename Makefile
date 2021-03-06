build:
	go build -o bin/socketbot bitbucket.org/pav5000/socketbot/cmd/bot

docker: build
	sudo docker build -t socketbot .

run:
	HTTP_PROXY=192.168.2.1:3128 HTTPS_PROXY=192.168.2.1:3128 go run bitbucket.org/pav5000/socketbot/cmd/bot
#	sudo docker run \
#		--rm -ti \
#		-v `pwd`/data:/bot/data  \
#		-e "HTTP_PROXY=192.168.2.1:3128" \
#		-e HTTPS_PROXY=192.168.2.1:3128  \
#		socketbot

start:
	sudo docker run -d --name=socketbot --restart=always -v `pwd`/data:/bot/data socketbot

stop:
	sudo docker rm -f socketbot; true

restart: stop start
	echo "restarted"

deploy:
	ssh pavl.uk -t 'bash -l -c "cd ~/docker/socketbot && git pull && make docker && make restart"'
