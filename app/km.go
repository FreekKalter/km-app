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

type TimeRow struct {
	Id                int64
	Date              time.Time
	CheckIn, CheckOut string
	Hours             float64
}
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
		fmt.Println("init:", err)
		slog.Fatal("init:", err)
	}
	var Dbmap *gorp.DbMap
	Dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	Dbmap.AddTable(Kilometers{}).SetKeys(true, "Id")
	Dbmap.AddTable(Times{}).SetKeys(false, "Id")

	var templates *template.Template
	if config.Env == "testing" {
		Dbmap.TraceOn("[gorp]", slog)
	} else {
		templates = template.Must(template.ParseFiles("index.html"))
	}
	s := &Server{Dbmap: Dbmap, templates: templates, log: slog, config: config}

	// static files get served directly
	if config.Env == "testing" {
		s.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("js/"))))
		s.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("img/"))))
		s.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))
		s.PathPrefix("/partials/").Handler(http.StripPrefix("/partials/", http.FileServer(http.Dir("partials/"))))
		s.Handle("/favicon.ico", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "favicon.ico") }))
	}

	s.HandleFunc("/", s.homeHandler).Methods("GET")
	s.HandleFunc("/state/{id}", s.stateHandler).Methods("GET")
	s.HandleFunc("/save/kilometers/{id}", s.saveKilometersHandler).Methods("POST")
	s.HandleFunc("/save/times/{id}", s.saveTimesHandler).Methods("POST")
	s.HandleFunc("/overview/{category}/{year}/{month}", s.overviewHandler).Methods("GET")
	s.HandleFunc("/delete/{id}", s.deleteHandler).Methods("GET")
	s.HandleFunc("/csv/{year}/{month}", s.csvHandler).Methods("GET")
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

func timeStamp(s *Server, action string, kilometersId int64) {
	id, err := s.Dbmap.SelectInt("select Id from times where date=$1", getDateStr())
	if err != nil {
		s.log.Println("timestamp:", err)
		return
	}
	today := new(Times)
	now := time.Now().Unix()
	if id == 0 { // no times saved for today
		today.Date = time.Now()
		today.Id = kilometersId
		s.log.Println("timestamp:", kilometersId, today.Id)
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
		Km, Time int
		Editable bool
	}
	type State struct {
		Begin, Eerste, Laatste, Terug Field
		LastDayError                  string
	}
	var state State

	// Get data save for this day
	var today Kilometers
	var err error
	if id == "today" {
		err = s.Dbmap.SelectOne(&today, "select * from kilometers where date=$1", dateStr)
	} else {
		if _, err := strconv.ParseInt(id, 10, 64); err != nil {
			http.Error(w, InvalidId.String(), InvalidId.Code)
			return
		}
		err = s.Dbmap.SelectOne(&today, "select * from kilometers where id=$1", id)
	}
	s.log.Println(err)
	switch {
	case err != nil && err.Error() != "sql: no rows in result set":
		http.Error(w, "Database error", 500)
		s.log.Println("stateHandler:", err)
		return
	case err != nil && err.Error() == "sql: no rows in result set": // today not saved yet TODO: check this
		s.log.Println("no today")
		var lastDay Kilometers
		err := s.Dbmap.SelectOne(&lastDay, "select * from kilometers where date = (select max(date) as date from kilometers)")
		if err != nil {
			http.Error(w, "Database error", 500)
			s.log.Println("stateHandler:", err)
			return
		}
		if lastDay != (Kilometers{}) { // Nothing in db yet
			state.Begin = Field{Km: lastDay.getMax(), Time: 0, Editable: true}
			state.Eerste = Field{Km: 0, Time: 0, Editable: true}
			state.Laatste = Field{Km: 0, Time: 0, Editable: true}
			state.Terug = Field{Km: 0, Time: 0, Editable: true}
		}

		var lastDayTimes Times
		err = s.Dbmap.SelectOne(&lastDayTimes, "select * from times where id=$1", lastDay.Id)
		if err == nil {
			if lastDayTimes.CheckIn == 0 || lastDayTimes.CheckOut == 0 {
				state.LastDayError = fmt.Sprintf("overview/tijden/%d/%d", lastDayTimes.Date.Year(), lastDayTimes.Date.Month())
			}
		}
	default: // Something is already filled in for today
		s.log.Println("today:", today)
		if today.Begin != 0 {
			state.Begin.Km = today.Begin
		} else {
			state.Begin.Editable = true
		}
		if today.Eerste == 0 {
			state.Eerste.Km = int(today.Begin / 1000)
			state.Eerste.Editable = true
		} else {
			state.Eerste.Km = today.Eerste
		}
		if today.Laatste == 0 {
			state.Laatste.Km = int(today.Eerste / 1000)
			state.Laatste.Editable = true
		} else {
			state.Laatste.Km = today.Laatste
		}
		if today.Terug == 0 {
			state.Terug.Km = int(today.Laatste / 1000)
			state.Terug.Editable = true
		} else {
			state.Terug.Km = today.Terug
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
		http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
		s.log.Println("overview:", err)
		return
	}
	year, err := strconv.ParseInt(vars["year"], 10, 64)
	if err != nil {
		http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
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
			http.Error(w, DbError.String(), DbError.Code)
			s.log.Println("overview:", err)
			return
		}
		jsonEncoder.Encode(all)
	case "tijden":
		rows, err := getAllTimes(s, year, month)
		if err != nil {
			http.Error(w, DbError.String(), DbError.Code)
			s.log.Println("overview:", err)
		}
		jsonEncoder.Encode(rows)
	default:
		http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
		return
	}
}

func getAllTimes(s *Server, year, month int64) (rows []TimeRow, err error) {
	var all []Times
	rows = make([]TimeRow, 0)
	_, err = s.Dbmap.Select(&all, "select * from times where extract (year from date)=$1 and extract (month from date)=$2 order by date desc ", year, month)
	if err != nil {
		return rows, err
	}
	loc, err := time.LoadLocation("Europe/Amsterdam") // should not be hardcoded but idgaf
	if err != nil {
		s.log.Println(err)
	}
	for _, c := range all {
		var row TimeRow
		row.Id = c.Id
		row.Date = c.Date
		if c.CheckIn != 0 {
			row.CheckIn = time.Unix(c.CheckIn, 0).In(loc).Format("15:04")
		} else {
			row.CheckIn = "-"
		}
		if c.CheckOut != 0 {
			row.CheckOut = time.Unix(c.CheckOut, 0).In(loc).Format("15:04")
		} else {
			row.CheckOut = "-"
		}
		if hours := (time.Duration(c.CheckOut-c.CheckIn) * time.Second).Hours(); hours > 0 && hours < 24 {
			row.Hours = hours
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (s *Server) csvHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	month, err := strconv.ParseInt(vars["month"], 10, 64)
	if err != nil {
		http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
		s.log.Println("overview:", err)
		return
	}
	year, err := strconv.ParseInt(vars["year"], 10, 64)
	if err != nil {
		http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
		s.log.Println("overview:", err)
		return
	}
	times, err := getAllTimes(s, year, month)
	if err != nil {
		http.Error(w, DbError.String(), DbError.Code)
		s.log.Println("overview:", err)
	}
	for _, t := range times {
		fmt.Fprintf(w, "%s,%s,%s,%.1f\n", t.Date.Format("Mon 2"), t.CheckIn, t.CheckOut, t.Hours)
	}
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	s.log.Println("delete:", id)
	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || intId < 0 {
		http.Error(w, InvalidId.String(), InvalidId.Code)
		return
	}
	var k = &Kilometers{Id: intId}
	var t = &Times{Id: intId}

	//_, err = s.Dbmap.Exec("delete from kilometers where id=$1", id)
	deleted, err := s.Dbmap.Delete(k)
	if err != nil || deleted == 0 {
		http.Error(w, InvalidId.String(), InvalidId.Code)
		return
	}
	s.log.Println("delete:", deleted, "from kilometers")
	deleted, err = s.Dbmap.Delete(t)
	if err != nil {
		s.log.Println("error deleting from times", err)
		http.Error(w, InvalidId.String(), InvalidId.Code)
		return
	}
	s.log.Println("delete:", err)
	fmt.Fprintf(w, "%d\n", intId)
}

func getDateStr() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d-%d-%d", now.Month(), now.Day(), now.Year())
}

func (s *Server) saveKilometersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// parse posted data into PostValue datastruct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, NotParsable.String(), NotParsable.Code)
		s.log.Println(err)
		return
	}
	var pv PostValue
	err = json.Unmarshal(body, &pv)
	if err != nil {
		http.Error(w, NotParsable.String(), NotParsable.Code)
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
		http.Error(w, UnknownField.String(), UnknownField.Code)
		return
	}

	today := new(Kilometers)
	s.log.Println("initialized today struct:", today)
	var id int64
	if vars["id"] == "today" { // use own dateStr
		now := time.Now().UTC()
		dateStr := fmt.Sprintf("%d-%d-%d", now.Month(), now.Day(), now.Year())
		err := s.Dbmap.SelectOne(today, "select * from kilometers where date=$1", dateStr)
		s.log.Println(err, today)
		if err == nil { // there is already data for today (so use update)
			s.log.Println(err, "dus selectONe goed gegaan, dus al data voor vandaag", today)
			today.addPost(pv)
			s.log.Println("na toevoegen van geposte data", today)
			_, err = s.Dbmap.Update(today)
			if err != nil {
				http.Error(w, DbError.String(), DbError.Code)
				s.log.Println(err)
				return
			}
		} else {
			s.log.Println("selectone gaf error, dus geen data voor vandaag")
			today = new(Kilometers) //reinit today, check if it was in database cleared this var
			today.Date = time.Now().UTC()
			today.addPost(pv)
			s.log.Println("hele struct die geinsert gaat worden", today)
			err = s.Dbmap.Insert(today)
			if err != nil {
				http.Error(w, DbError.String(), DbError.Code)
				s.log.Println(err)
				return
			}
		}
		id = today.Id
	} else { // id provided (so already an entry for sure), get it, add it and save it
		id, err = strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			http.Error(w, NotParsable.String(), NotParsable.Code)
			s.log.Println(err)
			return
		}
		err = s.Dbmap.SelectOne(today, "select * from kilometers where id=$1", id)
		if err != nil {
			http.Error(w, DbError.String(), DbError.Code)
			s.log.Println(err)
			return
		}
		today.addPost(pv)
		_, err = s.Dbmap.Update(today)
		if err != nil {
			http.Error(w, DbError.String(), DbError.Code)
			s.log.Println(err)
			return
		}
	}

	pv.Name = strings.ToLower(pv.Name)
	if pv.Name == "eerste" {
		go timeStamp(s, "in", id)
	} else if pv.Name == "laatste" {
		go timeStamp(s, "out", id)
	} else if pv.Name == "begin" {
		var days []Kilometers
		_, err := s.Dbmap.Select(&days, "select * from kilometers order by date desc limit 2")
		if err != nil {
			s.log.Println("erronr in saving yesterday", err)
		}
		if len(days) == 2 && days[1].Terug == 0 {
			days[1].Terug = pv.Value
			_, err := s.Dbmap.Update(&days[1])
			if err != nil {
				s.log.Println("trying to save yesterday:", err)
			}
		}

	}
	fmt.Fprintf(w, "%d", id)
}

func (s *Server) saveTimesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, NotParsable.String(), NotParsable.Code)
		s.log.Println(err)
		return
	}
	type TimesPost struct {
		Date, CheckIn, CheckOut string
	}
	var tp TimesPost
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, NotParsable.String(), NotParsable.Code)
		s.log.Println(err)
		return
	}
	s.log.Println("request body:", body)
	err = json.Unmarshal(body, &tp)
	if err != nil {
		http.Error(w, NotParsable.String(), NotParsable.Code)
		s.log.Println(err)
		return
	}
	s.log.Printf("unmarshalled req: %+v", tp)

	loc, err := time.LoadLocation("Europe/Amsterdam") // should not be hardcoded but idgaf
	var t Times
	err = s.Dbmap.SelectOne(&t, "select * from times where id=$1", id)
	if err != nil {
		http.Error(w, DbError.String(), DbError.Code)
		s.log.Println(err)
		return
	}

	if tp.CheckIn == "-" {
		t.CheckIn = 0
	} else {
		checkin, err := time.ParseInLocation("2-1-2006 15:04", fmt.Sprintf("%s %s", tp.Date, tp.CheckIn), loc)
		if err != nil {
			http.Error(w, NotParsable.String(), NotParsable.Code)
			s.log.Println(err)
			return
		}
		t.CheckIn = checkin.UTC().Unix()
	}

	if tp.CheckOut == "-" {
		t.CheckOut = 0
	} else {
		checkout, err := time.ParseInLocation("2-1-2006 15:04", fmt.Sprintf("%s %s", tp.Date, tp.CheckOut), loc)
		//checkout, err := time.Parse("2-1-2006 15:04", fmt.Sprintf("%s %s", tp.Date, tp.CheckOut))
		if err != nil {
			http.Error(w, NotParsable.String(), NotParsable.Code)
			s.log.Println(err)
			return
		}
		t.CheckOut = checkout.UTC().Unix()
	}
	_, err = s.Dbmap.Update(&t)
	if err != nil {
		http.Error(w, DbError.String(), DbError.Code)
		s.log.Println(err)
		return
	}
	fmt.Fprintf(w, "%d", id)
}
