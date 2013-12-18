package main

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"syscall"

	"net/http"

	"github.com/coopernurse/gorp"
	httpgzip "github.com/daaku/go.httpgzip"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var (
	dbmap *gorp.DbMap
	slog  *log.Logger
)

func main() {
	defer dbmap.Db.Close()
	r := mux.NewRouter()

	// static files get served directly
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", cacheHandler(http.FileServer(http.Dir("js/")), 30)))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", cacheHandler(http.FileServer(http.Dir("img/")), 30)))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cacheHandler(http.FileServer(http.Dir("css/")), 30)))
	r.PathPrefix("/partials/").Handler(http.StripPrefix("/partials/", cacheHandler(http.FileServer(http.Dir("partials/")), 30)))

	r.Handle("/favicon.ico", cacheHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "favicon.ico") }), 40))

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/state/{id}", stateHandler).Methods("GET")
	r.HandleFunc("/save", saveHandler).Methods("POST")
	r.HandleFunc("/overview/{category}/{year}/{month}", overviewHandler).Methods("GET")
	r.HandleFunc("/delete/{id}", deleteHandler).Methods("GET")

	http.Handle("/", r)
	slog.Println("started...")

	// wrap the whole mux router wich implements http.Handler in a gzip middleware
	http.ListenAndServe(":4001", httpgzip.NewHandler(r))
}

type PostValue struct {
	Name  string
	Value int
}

type Kilometers struct {
	Id                            int64
	Date                          time.Time
	Begin, Eerste, Laatste, Terug int
	Comment                       string
}

func (k *Kilometers) getMax() int {
	if k.Terug > 0 {
		return k.Terug
	}
	if k.Laatste > 0 {
		return k.Laatste
	}
	if k.Eerste > 0 {
		return k.Eerste
	}
	if k.Begin > 0 {
		return k.Begin
	}
	return 0
}

func (k *Kilometers) addPost(pv PostValue) {
	switch pv.Name {
	case "begin":
		k.Begin = pv.Value
	case "eerste":
		k.Eerste = pv.Value
	case "laatste":
		k.Laatste = pv.Value
	case "terug":
		k.Terug = pv.Value
	}
}

type Times struct {
	Id       int64
	Date     time.Time
	CheckIn  int64
	CheckOut int64
}

func timeStamp(action string) {
	id, err := dbmap.SelectInt("select Id from times where date=$1", getDateStr())
	if err != nil {
		slog.Fatal(err)
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
		dbmap.Insert(today)
	} else {
		err = dbmap.SelectOne(today, "select * from times where id=$1", id)
		if err != nil {
			slog.Fatal(err)
		}
		switch action {
		case "in":
			today.CheckIn = now
		case "out":
			today.CheckOut = now
		}
		dbmap.Update(today)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	indexContent, err := ioutil.ReadFile("index.html")
	if err != nil {
		slog.Fatal(err)
	}
	fmt.Fprintf(w, string(indexContent))
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
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
		err = dbmap.SelectOne(&today, "select * from kilometers where date=$1", dateStr)
	} else {
		err = dbmap.SelectOne(&today, "select * from kilometers where id=$1", id)
	}
	switch {
	case err != nil:
		if err != nil {
			slog.Fatal(err)
		}
	case today == (Kilometers{}): // today not saved yet TODO: check this
		slog.Println("no today")
		var lastDay Kilometers
		err := dbmap.SelectOne(&lastDay, "select * from kilometers where date = (select max(date) as date from kilometers)")
		if err != nil {
			slog.Fatal(err)
		}
		if lastDay == (Kilometers{}) { // Nothing in db yet

		}
		state.Begin = Field{lastDay.getMax(), true}
		state.Eerste = Field{0, true}
		state.Laatste = Field{0, true}
		state.Terug = Field{0, true}
	default: // Something is already filled in for today
		slog.Println(today)
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

func allowedMethod(method string) bool {
	for _, v := range []string{"Begin", "Eerste", "Laatste", "Terug"} {
		if v == method {
			return true
		}
	}
	return false
}

func overviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]
	month, err := strconv.ParseInt(vars["month"], 10, 64)
	if err != nil {
		slog.Fatal(err)
	}
	year, err := strconv.ParseInt(vars["year"], 10, 64)
	if err != nil {
		slog.Fatal(err)
	}
	slog.Println("overview", year, month)

	switch category {
	case "kilometers":
		all := make([]Kilometers, 0)
		_, err := dbmap.Select(&all, "select * from kilometers where extract (year from date)=$1 and extract (month from date)=$2 order by date desc ", year, month)
		if err != nil {
			slog.Fatal("overview:", err)
		}

		jsonEncoder := json.NewEncoder(w)
		jsonEncoder.Encode(all)

	case "tijden":
		var all []Times
		type Column struct {
			Date, CheckIn, CheckOut time.Time
			Hours                   int
		}
		columns := make([]Column, 0)
		_, err := dbmap.Select(&all, "select * from times where extract (year from date)=$1 and extract (month from date)=$2 order by date desc ", year, month)
		if err != nil {
			slog.Fatal(err)
		}
		for _, c := range all {
			var col Column
			col.Date = c.Date
			col.CheckIn = time.Unix(c.CheckIn, 0)
			col.CheckOut = time.Unix(c.CheckOut, 0)
			col.Hours = int((time.Duration(c.CheckOut-c.CheckIn) * time.Second).Hours())
			columns = append(columns, col)
		}
		jsonEncoder := json.NewEncoder(w)
		jsonEncoder.Encode(columns)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	slog.Println("delete:", id)
	_, err := dbmap.Exec("delete from kilometers where id=$1", id)
	if err != nil {
		slog.Fatal(err)
	}
}

func getDateStr() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d-%d-%d", now.Month(), now.Day(), now.Year())
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	// parse posted data into PostValue datastruct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		slog.Fatal(err)
	}
	var pv PostValue
	err = json.Unmarshal(body, &pv)
	if err != nil {
		slog.Fatal(err)
	}

	// sanitize columns
	if !allowedMethod(pv.Name) {
		return
	} else {
		pv.Name = strings.ToLower(pv.Name)
	}
	if pv.Name == "eerste" {
		go timeStamp("in")
	} else if pv.Name == "laatste" {
		go timeStamp("out")
	}

	dateStr := getDateStr()
	id, err := dbmap.SelectInt("select id from kilometers where date=$1", dateStr)

	today := new(Kilometers)
	if id == 0 { // nothing saved for today, save posted data and date
		today.Date = time.Now().UTC()
		today.addPost(pv)
		slog.Println(today)
		err = dbmap.Insert(today)
		if err != nil {
			slog.Fatal(err)
		}
	} else { // update already partially saved day
		err = dbmap.SelectOne(today, "select * from kilometers where id=$1", id)
		//today, err = dbmap.Get(Kilometers{}, id)
		if err != nil {
			slog.Fatal(err)
		}
		today.addPost(pv)
		_, err = dbmap.Update(today)
		if err != nil {
			slog.Fatal(err)
		}
	}

	fmt.Fprintf(w, "ok")
}

func init() {
	os.Chdir("/app")
	// Set up logging
	var err error
	logFile, err := os.OpenFile("/log/km.log", syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CREAT, 0666)
	slog = log.New(logFile, "km: ", log.LstdFlags)
	if err != nil {
		log.Panic(err)
	}

	db, err := sql.Open("postgres", "user=docker dbname=km password=docker sslmode=disable")
	if err != nil {
		slog.Fatal("dberror: ", err)
	}
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTable(Kilometers{}).SetKeys(true, "Id")
	dbmap.AddTable(Times{}).SetKeys(true, "Id")

	dbmap.TraceOn("[gorp]", log.New(logFile, "myapp:", log.Lmicroseconds))
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
