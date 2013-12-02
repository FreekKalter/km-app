all: app docker

app: app/km.go
	go build -o app/km app/km.go

docker: app
	sudo docker build -t freekkalter/km:test .
	sudo docker kill testing
	sudo docker rm testing
	sudo docker run -d -p 4001:4001 -v /home/fkalter/github/km/log:/log:rw -name testing freekkalter/km:test
