package main

import (
	"database/sql"
	"encoding/json"
	"net"

	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/fcgi"

	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/gorilla/mux"
	"launchpad.net/goyaml"

	_ "github.com/lib/pq"
)

type Server struct {
	mux.Router
	Dbmap     *gorp.DbMap
	templates *template.Template
	log       *log.Logger
	config    Config
}

func NewServer(dbName string, config Config) *Server {
	// Set up logging
	var slog *log.Logger
	if config.Log == "" {
		slog = log.New(ioutil.Discard, "km: ", log.LstdFlags)
	} else {
		logFile, err := os.OpenFile(config.Log, syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CREAT, 0666)
		if err != nil {
			log.Panic(err)
		}
		slog = log.New(logFile, "km: ", log.LstdFlags)
	}

	db, err := sql.Open("postgres", "user=docker dbname="+dbName+" password=docker sslmode=disable")
	if err != nil {
		slog.Fatal("init:", err)
	}
	var Dbmap *gorp.DbMap
	Dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	Dbmap.AddTable(Kilometers{}).SetKeys(true, "Id")
	Dbmap.AddTable(Times{}).SetKeys(true, "Id")

	var templates *template.Template
	if config.Env == "testing" {
		Dbmap.TraceOn("[gorp]", slog)
	} else {
		templates = template.Must(template.ParseFiles("index.html"))
	}
	s := &Server{Dbmap: Dbmap, templates: templates, log: slog, config: config}

	// static files get served directly
	s.PathPrefix("/js/").Handler(http.StripPrefix("/js/", cacheHandler(http.FileServer(http.Dir("js/")), 30)))
	s.PathPrefix("/img/").Handler(http.StripPrefix("/img/", cacheHandler(http.FileServer(http.Dir("img/")), 30)))
	s.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cacheHandler(http.FileServer(http.Dir("css/")), 30)))
	s.PathPrefix("/partials/").Handler(http.StripPrefix("/partials/", cacheHandler(http.FileServer(http.Dir("partials/")), 30)))

	s.Handle("/favicon.ico", cacheHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "favicon.ico") }), 40))

	s.HandleFunc("/", s.homeHandler).Methods("GET")
	s.HandleFunc("/state/{id}", s.stateHandler).Methods("GET")
	s.HandleFunc("/save", s.saveHandler).Methods("POST")
	s.HandleFunc("/overview/{category}/{year}/{month}", s.overviewHandler).Methods("GET")
	s.HandleFunc("/delete/{id}", s.deleteHandler).Methods("GET")
	return s
}

func main() {
	// Load config
	configFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		panic(err)
	}
	var config Config
	err = goyaml.Unmarshal(configFile, &config)
	if err != nil {
		panic(err)
	}

	s := NewServer("km", config)
	defer s.Dbmap.Db.Close()

	http.Handle("/", s)
	s.log.Printf("started... (%s)\n", config.Env)

	listener, _ := net.Listen("tcp", ":4001")
	if config.Env == "testing" {
		http.Serve(listener, nil)
	} else {
		fcgi.Serve(listener, nil)
	}
}

type Config struct {
	Env string
	Log string
}

type PostValue struct {
	Name  string
	Value int
}

type Times struct {
	Id       int64
	Date     time.Time
	CheckIn  int64
	CheckOut int64
}

func timeStamp(s *Server, action string) {
	id, err := s.Dbmap.SelectInt("select Id from times where date=$1", getDateStr())
	if err != nil {
		s.log.Println("timestamp:", err)
		return
	}
	today := new(Times)
	now := time.Now().Unix()
	if id == 0 { // no times saved for today
		today.Date = time.Now()
		switch action {
		case "in":
			today.CheckIn = now
		case "out":
			today.CheckOut = now
		}
		s.Dbmap.Insert(today)
	} else {
		err = s.Dbmap.SelectOne(today, "select * from times where id=$1", id)
		if err != nil {
			s.log.Println("timestamp:", err)
			return
		}
		switch action {
		case "in":
			today.CheckIn = now
		case "out":
			today.CheckOut = now
		}
		s.Dbmap.Update(today)
	}
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if s.config.Env == "testing" {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, s.config)
	} else {
		s.templates.Execute(w, s.config)
	}
}

func (s *Server) stateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	now := time.Now().UTC()
	dateStr := fmt.Sprintf("%d-%d-%d", now.Month(), now.Day(), now.Year())
	type Field struct {
		Value    int
		Editable bool
	}
	type State struct {
		Date                          string
		Begin, Eerste, Laatste, Terug Field
	}
	var state State
	state.Date = dateStr

	// Get data save for this day
	var today Kilometers
	var err error
	if id == "today" {
		err = s.Dbmap.SelectOne(&today, "select * from kilometers where date=$1", dateStr)
	} else {
		if _, err := strconv.ParseInt(id, 10, 64); err != nil {
			http.Error(w, InvalidId.Body, InvalidId.Code)
			return
		}

		err = s.Dbmap.SelectOne(&today, "select * from kilometers where id=$1", id)
	}
	switch {
	case err != nil:
		if err != nil {
			http.Error(w, "Database error", 500)
			s.log.Println("stateHandler:", err)
			return
		}
	case today == (Kilometers{}): // today not saved yet TODO: check this
		s.log.Println("no today")
		var lastDay Kilometers
		err := s.Dbmap.SelectOne(&lastDay, "select * from kilometers where date = (select max(date) as date from kilometers)")
		if err != nil {
			http.Error(w, "Database error", 500)
			s.log.Println("stateHandler:", err)
			return
		}
		if lastDay == (Kilometers{}) { // Nothing in db yet

		}
		state.Begin = Field{lastDay.getMax(), true}
		state.Eerste = Field{0, true}
		state.Laatste = Field{0, true}
		state.Terug = Field{0, true}
	default: // Something is already filled in for today
		s.log.Println(today)
		if today.Begin != 0 {
			state.Begin.Value = today.Begin
		} else {
			state.Begin.Editable = true
		}
		if today.Eerste == 0 {
			state.Eerste.Value = int(today.Begin / 1000)
			state.Eerste.Editable = true
		} else {
			state.Eerste.Value = today.Eerste
		}
		if today.Laatste == 0 {
			state.Laatste.Value = int(today.Eerste / 1000)
			state.Laatste.Editable = true
		} else {
			state.Laatste.Value = today.Laatste
		}
		if today.Terug == 0 {
			state.Terug.Value = int(today.Laatste / 1000)
			state.Terug.Editable = true
		} else {
			state.Terug.Value = today.Terug
		}
	}
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(state)
}

func (s *Server) overviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]
	month, err := strconv.ParseInt(vars["month"], 10, 64)
	if err != nil {
		http.Error(w, InvalidUrl.Body, InvalidUrl.Code)
		s.log.Println("overview:", err)
		return
	}
	year, err := strconv.ParseInt(vars["year"], 10, 64)
	if err != nil {
		http.Error(w, InvalidUrl.Body, InvalidUrl.Code)
		s.log.Println("overview:", err)
		return
	}
	s.log.Println("overview", year, month)

	jsonEncoder := json.NewEncoder(w)
	switch category {
	case "kilometers":
		all := make([]Kilometers, 0)
		_, err := s.Dbmap.Select(&all, "select * from kilometers where extract (year from date)=$1 and extract (month from date)=$2 order by date desc ", year, month)
		if err != nil {
			http.Error(w, DbError.Body, DbError.Code)
			s.log.Println("overview:", err)
			return
		}
		jsonEncoder.Encode(all)
	case "tijden":
		var all []Times
		type Column struct {
			Date, CheckIn, CheckOut time.Time
			Hours                   float64
		}
		columns := make([]Column, 0)
		_, err := s.Dbmap.Select(&all, "select * from times where extract (year from date)=$1 and extract (month from date)=$2 order by date desc ", year, month)
		if err != nil {
			http.Error(w, DbError.Body, DbError.Code)
			s.log.Println("overview:", err)
			return
		}
		for _, c := range all {
			var col Column
			col.Date = c.Date
			col.CheckIn = time.Unix(c.CheckIn, 0)
			col.CheckOut = time.Unix(c.CheckOut, 0)
			col.Hours = (time.Duration(c.CheckOut-c.CheckIn) * time.Second).Hours()
			columns = append(columns, col)
		}
		jsonEncoder.Encode(columns)
	default:
		http.Error(w, InvalidUrl.Body, InvalidUrl.Code)
		return
	}
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	s.log.Println("delete:", id)
	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || intId < 0 {
		http.Error(w, InvalidId.Body, InvalidId.Code)
		return
	}
	var k = &Kilometers{Id: intId}

	//_, err = s.Dbmap.Exec("delete from kilometers where id=$1", id)
	deleted, err := s.Dbmap.Delete(k)
	if err != nil || deleted == 0 {
		http.Error(w, InvalidId.Body, InvalidId.Code)
		return
	}
	s.log.Println("delete:", err)
	fmt.Fprintf(w, "%d\n", intId)
}

func getDateStr() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d-%d-%d", now.Month(), now.Day(), now.Year())
}

func (s *Server) saveHandler(w http.ResponseWriter, r *http.Request) {
	// parse posted data into PostValue datastruct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, NotParsable.Body, NotParsable.Code)
		s.log.Println(err)
		return
	}
	var pv PostValue
	err = json.Unmarshal(body, &pv)
	if err != nil {
		http.Error(w, NotParsable.Body, NotParsable.Code)
		s.log.Println(err)
		return
	}

	// sanitize columns
	san := func(method string) bool {
		for _, v := range []string{"Begin", "Eerste", "Laatste", "Terug"} {
			if method == v {
				return true
			}
		}
		return false
	}
	if !san(pv.Name) {
		http.Error(w, UnknownField.Body, UnknownField.Code)
		return
	}
	pv.Name = strings.ToLower(pv.Name)
	if pv.Name == "eerste" {
		go timeStamp(s, "in")
	} else if pv.Name == "laatste" {
		go timeStamp(s, "out")
	}

	dateStr := getDateStr()
	id, err := s.Dbmap.SelectInt("select id from kilometers where date=$1", dateStr)

	//today := new(Kilometers)
	today := NewKilometers()
	if id == 0 { // nothing saved for today, save posted data and date
		today.addPost(pv)
		s.log.Println(today)
		err = s.Dbmap.Insert(today)
		if err != nil {
			http.Error(w, DbError.Body, DbError.Code)
			s.log.Println(err)
			return
		}
	} else { // update already partially saved day
		err = s.Dbmap.SelectOne(today, "select * from kilometers where id=$1", id)
		//today, err = s.Dbmap.Get(Kilometers{}, id)
		if err != nil {
			http.Error(w, DbError.Body, DbError.Code)
			s.log.Println(err)
			return
		}
		today.addPost(pv)
		_, err = s.Dbmap.Update(today)
		if err != nil {
			http.Error(w, DbError.Body, DbError.Code)
			s.log.Println(err)
			return
		}
	}
	fmt.Fprintf(w, "ok\n")
}

func cacheHandler(h http.Handler, days int) http.Handler {
	dur := time.Duration(days) * time.Duration(24) * time.Hour
	// ourHandler implements the http.Handler interface
	ourHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", int64(dur.Seconds())))
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(ourHandler)
}
