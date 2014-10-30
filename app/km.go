package main

import (
	"io/ioutil"
	"net"
	"net/http"

	"log"

	"bitbucket.org/FreekKalter/km"
	"launchpad.net/goyaml"
)

func main() {
	// Load config
	configFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		panic(err)
	}
	var config km.Config
	err = goyaml.Unmarshal(configFile, &config)
	if err != nil {
		panic(err)
	}
	s := km.NewServer("km", config)
	defer s.Dbmap.Db.Close()

	http.Handle("/", s)
	log.Printf("started... (%s)\n", config.Env)

	listener, _ := net.Listen("tcp", ":4001")
	if config.Env == "testing" {
		http.Serve(listener, nil)
	} else {
		http.Serve(listener, nil)
		//fcgi.Serve(listener, nil)
	}
}
