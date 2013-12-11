package main

import (
	"database/sql"
	"encoding/json"

	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"syscall"

	"net/http"

	"github.com/coopernurse/gorp"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

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

func (k Kilometers) getMax() int {
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
		slog.Println("updating postdata in today")
	case "eerste":
		k.Eerste = pv.Value
	case "laatste":
		k.Laatste = pv.Value
	case "terug":
		k.Terug = pv.Value
	}
}

var dbmap *gorp.DbMap
var slog *log.Logger

func homeHandler(w http.ResponseWriter, r *http.Request) {
	indexContent, err := ioutil.ReadFile("index.html")
	if err != nil {
		slog.Fatal(err)
	}
	fmt.Fprintf(w, string(indexContent))
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
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
	err := dbmap.SelectOne(&today, "select * from kilometers where date=$1", dateStr)
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
	slog.Println("get overview")
	var all []Kilometers
	_, err := dbmap.Select(&all, "select * from kilometers order by date")
	if err != nil {
		slog.Fatal("overview:", err)
	}

	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(all)
	// TODO
	//datestring := date.Format("02-01-2006")
}

func getDateStr() string {
	now := time.Now()
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

	dateStr := getDateStr()
	id, err := dbmap.SelectInt("select id from kilometers where date=$1", dateStr)

	today := new(Kilometers)
	if id == 0 { // nothing saved for today, save posted data and date
		today.Date = time.Now()
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
	//logFile, err := os.Create("/log/km.log")
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

	dbmap.TraceOn("[gorp]", log.New(logFile, "myapp:", log.Lmicroseconds))

	//hd, err = hood.Open("postgres", "user=docker dbname=km password=docker sslmode=disable")
}

func main() {
	defer dbmap.Db.Close()
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/state", stateHandler)
	r.HandleFunc("/save", saveHandler).Methods("POST")
	r.HandleFunc("/overview", overviewHandler).Methods("GET")

	// static files get served directly
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("js/"))))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("img/"))))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))
	r.PathPrefix("/partials/").Handler(http.StripPrefix("/partials/", http.FileServer(http.Dir("partials/"))))

	http.Handle("/", r)
	slog.Println("started...")

	http.ListenAndServe(":4001", r)
}
