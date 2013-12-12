package main

import (
	"flag"

	"log"
	"math/rand"
	"time"

	"database/sql"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

var dbmap *gorp.DbMap
var startDate, endDate string
var km int

type Kilometers struct {
	Id                            int64
	Date                          time.Time
	Begin, Eerste, Laatste, Terug int
	Comment                       string
}

func init() {
	db, err := sql.Open("postgres", "user=docker dbname=km password=docker sslmode=disable")
	if err != nil {
		log.Fatal("dberror: ", err)
	}
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTable(Kilometers{}).SetKeys(true, "Id")

	rand.Seed(time.Now().UnixNano())
	flag.StringVar(&startDate, "from", "01-11-2013", "start date")
	flag.StringVar(&endDate, "to", "01-12-2013", "end date")
	flag.IntVar(&km, "start", 14000, "number of kilometers to start with")
	flag.Parse()
}
func main() {
	const dateFormat = "02-01-2006"
	start, err := time.Parse(dateFormat, startDate)
	if err != nil {
		log.Fatal("startDate:", err)
	}
	end, err := time.Parse(dateFormat, endDate)
	if err != nil {
		log.Fatal("endDate:", err)
	}
	if start.After(end) {
		log.Fatal("start must be before end")
	}

	for start.Before(end) {
		k := new(Kilometers)
		k.Date = start
		k.Begin = km
		km += rand.Intn(50)
		k.Eerste = km
		km += rand.Intn(100)
		k.Laatste = km
		km += rand.Intn(20)
		k.Terug = km

		err := dbmap.Insert(k)
		if err != nil {
			log.Println(err)
		}
		//start = start.Add(time.Duration(rand.Intn(5)) * time.Duration(24) * time.Hour)
		start = start.Add(time.Duration(24) * time.Hour)
	}
}
