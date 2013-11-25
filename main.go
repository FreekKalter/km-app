package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	indexContent, err := ioutil.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, string(indexContent))
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	dateStr := fmt.Sprintf("%d-%d-%d", now.Month(), now.Day(), now.Year())
	var (
		id, begin, arnhem, laatste, terugkomst int
	)

	err := db.QueryRow("SELECT id,begin,arnhem,laatste,terugkomst FROM km WHERE date=$1", dateStr).Scan(&id, &begin, &arnhem, &laatste, &terugkomst)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	state := map[string]interface{}{
		"date":       dateStr,
		"begin":      begin,
		"arnhem":     arnhem,
		"laatste":    laatste,
		"terugkomst": terugkomst,
	}
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(state)
}

func allowedMethod(method string) bool {
	for _, v := range []string{"begin", "arnhem", "laatste", "terugkomst"} {
		if v == method {
			return true
		}
	}
	return false
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	// parse posted data into PostValue datastruct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	var pv PostValue
	err = json.Unmarshal(body, &pv)
	if err != nil {
		panic(err)
	}

	// sanitize columns
	if !allowedMethod(pv.Name) {
		return
	}

	var id int
	now := time.Now()
	dateStr := fmt.Sprintf("%d-%d-%d", now.Month(), now.Day(), now.Year())
	insert, _ := db.Prepare(fmt.Sprintf("insert into km (%s, date) values($1,$2)", pv.Name))
	update, _ := db.Prepare(fmt.Sprintf("update km set %s=$1 where id=$2", pv.Name))
	err = db.QueryRow("SELECT id FROM km WHERE date=$1", dateStr).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		_, err := insert.Exec(pv.Value, dateStr)
		if err != nil {
			panic(err)
		}
		fmt.Println("add")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Println("update")
		_, err := update.Exec(pv.Value, id)
		if err != nil {
			panic(err)
		}
	}

	fmt.Fprintf(w, "sure")
}

func init() {
	var err error
	db, err = sql.Open("postgres", "user=postgres dbname=postgres password=password sslmode=disable")
	if err != nil {
		panic(err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/state", stateHandler)
	r.HandleFunc("/save", saveHandler).Methods("POST")

	// static files get served directly
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("js/"))))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("img/"))))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))

	http.Handle("/", r)
	fmt.Println("started...")
	http.ListenAndServe(":4001", r)
}
