app/km: app/km.go
	go build -o app/km app/km.go

# TODO: find a way to keep the container running and just restart the app
# already running supervisor, so maybe also start a service wich listens for restart commands (might be overkill)
test-run: app/km
	docker kill `cat .cidfile`
	rm .cidfile
	docker run -cidfile=./.cidfile -v /home/fkalter/postgresdata:/data:rw\
								   -v /home/fkalter/github/km/app:/app:rw\
								   -v /home/fkalter/github/km/log:/log:rw\
								   -d -p 4001:4001 -p 5432:5432\
								   freekkalter/postgres-supervisord:km /usr/bin/supervisord
