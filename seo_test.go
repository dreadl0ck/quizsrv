package main

import (
	"encoding/xml"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchRoutes(t *testing.T) {
	previous := data
	defer func() { data = previous }()
	data = &Data{Courses: map[string]*course{"security": {}, "networking": {}}}
	w := httptest.NewRecorder()
	sitemap(w, httptest.NewRequest("GET", "/sitemap.xml", nil))
	var result struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.URLs) != 3 || result.URLs[1].Loc != "https://sne.dreadl0ck.net/courses/networking" {
		t.Fatal(w.Body.String())
	}
	w = httptest.NewRecorder()
	home(w, httptest.NewRequest("GET", "/missing-page", nil))
	if w.Code != 404 {
		t.Fatalf("unknown URL returned %d", w.Code)
	}
	w = httptest.NewRecorder()
	nobots(w, httptest.NewRequest("GET", "/robots.txt", nil))
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain") || strings.Contains(w.Body.String(), "Disallow: /\n") {
		t.Fatal("public catalog blocked or invalid robots response")
	}
}
