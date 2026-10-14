package short

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// A page behind a sign-in (403) still gets a short link, without details; a readable page gives its title.
func TestDetailsOrNone(t *testing.T) {
	private := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "sign in", http.StatusForbidden) }))
	defer private.Close()
	if d := detailsOrNone(private.URL); d == nil || d.Title != "" {
		t.Fatalf("private page: %+v", d)
	}
	public := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Tracklines</title></head></html>`))
	}))
	defer public.Close()
	if d := detailsOrNone(public.URL); d == nil || d.Title != "Tracklines" {
		t.Fatalf("public page: %+v", d)
	}
	if d := detailsOrNone("http://127.0.0.1:1/unreachable"); d == nil {
		t.Fatal("unreachable page: nil details")
	}
}

// A title from the caller is used without reading the page; without one, the page is read.
func TestDetailsGiven(t *testing.T) {
	read := false
	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		read = true
		_, _ = w.Write([]byte(`<html><head><title>From the page</title></head></html>`))
	}))
	defer page.Close()
	if d := details(page.URL, "Given", "About it"); d.Title != "Given" || d.Description != "About it" || read {
		t.Fatalf("given title: %+v, page read: %v", d, read)
	}
	if d := details(page.URL, "", ""); d.Title != "From the page" || !read {
		t.Fatalf("no title: %+v, page read: %v", d, read)
	}
}
