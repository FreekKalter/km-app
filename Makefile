
km: km.go
	go build km.go
	-pkill km
	./km &
