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
	config, err := parseConfig("config.yml")
	if err != nil {
		log.Fatal(err.Error())
	}

	s, err := km.NewServer("km", config)
	if err != nil {
		log.Fatal(err.Error())
	}
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

//TODO:test this function
func parseConfig(filename string) (config km.Config, err error) {
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = goyaml.Unmarshal(configFile, &config)
	if err != nil {
		return
	}
	return
}
