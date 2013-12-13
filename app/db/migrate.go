package main

import (
	"log"

	"time"

	"database/sql"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

var dbKm, dbKilometers *gorp.DbMap
var startDate, endDate string

type Kilometers struct {
	Id                            int64
	Date                          time.Time
	Begin, Eerste, Laatste, Terug int
	Comment                       string
}

type Km struct {
	Id                                 int64
	Date                               time.Time
	Begin, Arnhem, Laatste, Terugkomst int
	Comment                            string
}

func init() {
	db, err := sql.Open("postgres", "user=docker dbname=km password=docker sslmode=disable")
	if err != nil {
		log.Fatal("dberror: ", err)
	}
	dbKm = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbKm.AddTable(Km{}).SetKeys(true, "Id")

	db2, err := sql.Open("postgres", "user=docker dbname=km password=docker sslmode=disable")
	if err != nil {
		log.Fatal("dberror: ", err)
	}
	dbKilometers = &gorp.DbMap{Db: db2, Dialect: gorp.PostgresDialect{}}
	dbKilometers.AddTable(Kilometers{}).SetKeys(true, "Id")
}
func main() {
	var kms []Km
	_, err := dbKm.Select(&kms, "select * from km")
	if err != nil {
		log.Fatal(err)
	}
	for _, k := range kms {
		var kilo Kilometers
		kilo.Id = k.Id
		kilo.Date = k.Date
		kilo.Begin = k.Begin
		kilo.Eerste = k.Arnhem
		kilo.Laatste = k.Laatste
		kilo.Terug = k.Terugkomst
		dbKilometers.Insert(kilo)
	}
}
