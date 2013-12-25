package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

var config Config = Config{Env: "production", Log: ""}
var s *Server = NewServer("km_test", config)
var db *gorp.DbMap = s.Dbmap

type response struct {
	body string
	code int
}

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

		if b := w.Body.String(); !strings.Contains(b, resp.body) {
			t.Fatalf("body = %q, want '%s'", b, resp.body)
		}
		if w.Code != resp.code {
			t.Fatalf("code = %d, want %d", w.Code, resp.code)
		}
	}
}

type TestCombo struct {
	req  *http.Request
	resp response
}

func NewTestCombo(url string, resp response) *TestCombo {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	return &TestCombo{req, resp}
}

func TestDelete(t *testing.T) {
	clearTable(t, "kilometers")

	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/delete/1", response{"Unknown id.", 500}),
		NewTestCombo("/delete/a", response{"Invalid id.", 500}),
		NewTestCombo("/delete/-1", response{"Invalid id.", 500}),
	}
	tableDrivenTest(t, table)
}

func TestSave(t *testing.T) {
	clearTable(t, "kilometers")

	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/save/a", response{"404 page not found\n", 404}),
	}
	req, _ := http.NewRequest("POST", "/save", strings.NewReader(`{"Name": "Begin", "Value": 1234}`))
	table = append(table, &TestCombo{req, response{"ok", 200}})

	tableDrivenTest(t, table)
}
