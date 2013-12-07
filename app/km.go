package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"

	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type PostValue struct {
	Name  string
	Value int
}

var db *sql.DB
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
	var (
		id, begin, arnhem, laatste, terugkomst int
	)

	type Field struct {
		Value    int
		Editable bool
	}
	type State struct {
		Date                               string
		Begin, Arnhem, Laatste, Terugkomst Field
	}
	var state State
	state.Date = dateStr

	err := db.QueryRow("SELECT id,begin,arnhem,laatste,terugkomst FROM km WHERE date=$1", dateStr).Scan(&id, &begin, &arnhem, &laatste, &terugkomst)
	switch {
	case err == sql.ErrNoRows: // new row
		err := db.QueryRow(`select id, begin, arnhem,laatste, terugkomst
                            from km
                            inner join(
                                select max(date) date
                                from km ) kmi
                            on km.date = kmi.date
                            limit 1;`).Scan(&id, &begin, &arnhem, &laatste, &terugkomst)
		switch {
		case err == sql.ErrNoRows:
			state.Begin = Field{0, true}
			state.Arnhem = Field{0, true}
			state.Laatste = Field{0, true}
			state.Terugkomst = Field{0, true}
		case err != nil:
			slog.Fatal(err)

		default:
			state.Begin = Field{terugkomst, true}
			state.Arnhem = Field{int(terugkomst / 1000), true} // first 3 digits of last km
			state.Laatste = Field{int(terugkomst / 1000), true}
			state.Terugkomst = Field{int(terugkomst / 1000), true}
		}

	case err != nil:
		if err != nil {
			slog.Fatal(err)
		}
	default: // Something is already filled in for today
		if begin != 0 {
			state.Begin.Value = begin
		}
		if arnhem == 0 {
			state.Arnhem.Value = int(begin / 1000)
			state.Arnhem.Editable = true
		} else {
			state.Arnhem.Value = arnhem
		}
		if laatste == 0 {
			state.Laatste.Value = int(begin / 1000)
			state.Laatste.Editable = true
		} else {
			state.Laatste.Value = laatste
		}
		if terugkomst == 0 {
			state.Terugkomst.Value = int(begin / 1000)
			state.Terugkomst.Editable = true
		} else {
			state.Terugkomst.Value = terugkomst
		}
	}

	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(state)
}

func allowedMethod(method string) bool {
	for _, v := range []string{"Begin", "Arnhem", "Laatste", "Terugkomst"} {
		if v == method {
			return true
		}
	}
	return false
}
func overviewHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`select date, begin, arnhem, laatste, terugkomst from km`)
	if err != nil {
		slog.Fatal(err)
	}
	var (
		begin, arnhem, laatste, terugkomst int
		date                               time.Time
	)
	type Day struct {
		Date                               string
		Begin, Arnhem, Laatste, Terugkomst int
	}
	var days []Day
	for rows.Next() {
		slog.Println("processing row")
		if err := rows.Scan(&date, &begin, &arnhem, &laatste, &terugkomst); err != nil {
			slog.Fatal(err)
		}
		datestring := date.Format("02-01-2006")
		days = append(days, Day{datestring, begin, arnhem, laatste, terugkomst})
		slog.Println(days)
	}
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(days)
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

	var id int
	now := time.Now()
	dateStr := fmt.Sprintf("%d-%d-%d", now.Month(), now.Day(), now.Year())
	update, _ := db.Prepare(fmt.Sprintf("update km set %s=$1 where id=$2", pv.Name))
	err = db.QueryRow("SELECT id FROM km WHERE date=$1", dateStr).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		slog.Println("add")
		var lastId int
		err = db.QueryRow(`insert into km (begin,arnhem,laatste,terugkomst, comment, date)
                                  values(0,0,0,0, '',$1) returning id`, dateStr).Scan(&lastId)
		if err != nil {
			slog.Fatal(err)
		}

		_, err := update.Exec(pv.Value, lastId)
		if err != nil {
			slog.Fatal(err)
		}
	case err != nil:
		slog.Fatal(err)
	default:
		slog.Println("update")
		_, err := update.Exec(pv.Value, id)
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

	db, err = sql.Open("postgres", "user=docker dbname=km password=docker sslmode=disable")
	if err != nil {
		slog.Fatal("dberror: ", err)
	}
}

func main() {
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
