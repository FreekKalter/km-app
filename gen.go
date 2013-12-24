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

var (
	dbmap              *gorp.DbMap
	startDate, endDate string
	km                 int
)

const (
	month      = time.Hour * time.Duration(720)
	day        = time.Hour * time.Duration(24)
	dateFormat = "02-01-2006"
)

type Kilometers struct {
	Id                            int64
	Date                          time.Time
	Begin, Eerste, Laatste, Terug int
	Comment                       string
}

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.StringVar(&startDate, "from", time.Now().Add(-month).Format(dateFormat), "start date")
	flag.StringVar(&endDate, "to", time.Now().Add(-day).Format(dateFormat), "end date")
	flag.IntVar(&km, "start", 14000, "number of kilometers to start with")
	dbName := flag.String("db", "km", "database to use")
	flag.Parse()

	db, err := sql.Open("postgres", "user=docker dbname="+*dbName+" password=docker sslmode=disable")
	if err != nil {
		log.Fatal("dberror: ", err)
	}
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTable(Kilometers{}).SetKeys(true, "Id")
}
func main() {
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
