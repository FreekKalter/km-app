package main

import (
	"fmt"
	"time"
)

/// Save times
type Times struct {
	Id                                int64
	Date                              time.Time
	Begin, CheckIn, CheckOut, Laatste int64
}

func (t *Times) updateObject(s *Server, date string, fields []Field) error {
	loc, _ := time.LoadLocation("Europe/Amsterdam") // should not be hardcoded but idgaf
	for _, field := range fields {
		if field.Time == "" {
			continue
		}
		s.log.Println(field.Name, field.Time)
		fieldLocalTime, err := time.ParseInLocation("1-2-2006 15:04", fmt.Sprintf("%s %s", date, field.Time), loc)
		s.log.Println(fieldLocalTime)
		if err != nil {
			return err
		}
		fieldTime := fieldLocalTime.UTC().Unix()
		switch field.Name {
		case "Begin":
			t.Begin = fieldTime
		case "Eerste":
			t.CheckIn = fieldTime
		case "Laatste":
			t.CheckOut = fieldTime
		case "Terug":
			t.Laatste = fieldTime
		}
	}
	return nil
}
