package main

import (
	"encoding/json"

	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

var config Config = Config{Env: "production", Log: ""}
var s *Server = NewServer("km_test", config)
var db *gorp.DbMap = s.Dbmap

func clearTable(t *testing.T, tableName string) {
	_, err := db.Exec("truncate kilometers")
	if err != nil {
		t.Fatal("truncating db: ", err)
	}
	slice, err := db.Select(Kilometers{}, "select * from kilometers")
	if len(slice) > 0 {
		t.Errorf("expected empty kilometers table to start with")
	}
	if err != nil {
		t.Errorf("could not select from kilometers")
	}
}

func tableDrivenTest(t *testing.T, table []*TestCombo) {
	for _, tc := range table {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, tc.req)
		resp := tc.resp

		if b := w.Body.String(); !strings.Contains(b, resp.Body) {
			t.Fatalf("body = %q, want '%s'", b, resp.Body)
		}
		if w.Code != resp.Code {
			t.Fatalf("code = %d, want %d", w.Code, resp.Code)
		}
	}
}

type TestCombo struct {
	req  *http.Request
	resp Response
}

func NewTestCombo(url string, resp Response) *TestCombo {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	return &TestCombo{req, resp}
}

func TestDelete(t *testing.T) {
	clearTable(t, "kilometers")

	// add a row, save id
	k := Kilometers{Begin: 1234}
	err := db.Insert(&k)
	if err != nil {
		t.Fatal("TestDelete: dberror on insert", err)
	}
	id := strconv.FormatInt(k.Id, 10)

	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/delete/1", InvalidId),
		NewTestCombo("/delete/a", InvalidId),
		NewTestCombo("/delete/-1", InvalidId),
		// delete saved row, compare returned id
		NewTestCombo("/delete/"+id, Response{id + "\n", 200}),
	}
	tableDrivenTest(t, table)
}

func TestSave(t *testing.T) {
	clearTable(t, "kilometers")

	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/save", NotFound),
		NewTestCombo("/save/a", NotFound),
	}
	req, _ := http.NewRequest("POST", "/save", strings.NewReader(`{"Name": "Begin", "Value": 1234}`))
	table = append(table, &TestCombo{req, Ok})

	req, _ = http.NewRequest("POST", "/save", strings.NewReader(`{"Name": "Begin", "Value": "abc"}`))
	table = append(table, &TestCombo{req, NotParsable})

	req, _ = http.NewRequest("POST", "/save", strings.NewReader(`{"Name": "InvalidFieldname", "Value": 1234}`))
	table = append(table, &TestCombo{req, UnknownField})

	tableDrivenTest(t, table)
}

func TestHome(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("expected 200 got ", w.Code)
	}
}

func TestState(t *testing.T) {
	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/state", NotFound),
		NewTestCombo("/state/2234a", InvalidId),
	}
	tableDrivenTest(t, table)

	req, _ := http.NewRequest("GET", "/state/today", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("expected 200 got ", w.Code)
	}
	var unMarschalled interface{}
	err := json.Unmarshal(w.Body.Bytes(), &unMarschalled)
	if err != nil {
		t.Fatal("/state/today not a valid json response:", err)
	}
}

func TestOverview(t *testing.T) {
	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/overview", NotFound),
		NewTestCombo("/overview/invalidCategory/2013/01", InvalidUrl),
		NewTestCombo("/overview/tijden/201a/01", InvalidUrl),
		NewTestCombo("/overview/tijden/2013/0a", InvalidUrl),
		NewTestCombo("/overview/1/2013/01", InvalidUrl),
	}
	tableDrivenTest(t, table)

}
