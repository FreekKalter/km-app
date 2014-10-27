package main

import (
	"fmt"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	ti := Times{}
	if ti.Id != 0 {
		t.Fatal("id should be 0 on times init")
	}
}

func TestUpdate(t *testing.T) {
	testDate := time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)

	fields := []Field{Field{Km: 123456, Time: "13:00", Name: "Begin"}}
	cmpDate := int64(1257854400) // 10-11-2009 13:00 uur (in unix formaat)

	dateStr := fmt.Sprintf("%d-%d-%d", testDate.Month(), testDate.Day(), testDate.Year())
	ti := Times{}
	ti.updateObject(s, dateStr, fields)
	if ti.Begin != cmpDate {
		t.Fatalf("updating fields failed: expected %d got %d", cmpDate, ti.Begin)
	}

	if ti.CheckIn != 0 || ti.CheckOut != 0 || ti.Laatste != 0 {
		t.Fatalf("only update 1 field of testtime struct, all other fields should be 0 but there not %+v", ti)
	}
}
