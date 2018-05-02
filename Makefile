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

deploy:
	echo "Don't forget to commit before the deploy"
	ssh pavl.uk -t 'bash -l -c "cd go/src/github.com/pav5000/socketbot && git pull && make docker && make restart"'
