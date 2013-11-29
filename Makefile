
run: build
	-pkill km
	./km &

build: km.go
	go build km.go
