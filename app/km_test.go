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

func clearTable(tableName string) {
	db.Exec("truncate kilometers")
}

func TestDelete(t *testing.T) {
	clearTable("kilometers")
	slice, err := db.Select(Kilometers{}, "select * from kilometers")
	if len(slice) > 0 {
		t.Errorf("expected empty kilometers table to start with")
	}
	if err != nil {
		t.Errorf("could not select from kilometers")
	}
	var table map[string]response = map[string]response{
		"/delete/1":  response{"Unknown id.", 500},
		"/delete/a":  response{"Invalid id.", 500},
		"/delete/-1": response{"Invalid id.", 500},
	}
	for url, resp := range table {
		r, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)

		if b := w.Body.String(); !strings.Contains(b, resp.body) {
			t.Fatalf("body = %q, want '%s'", b, resp.body)
		}
		if w.Code != resp.code {
			t.Fatalf("code = %d, want %d", w.Code, resp.code)
		}
	}

}
