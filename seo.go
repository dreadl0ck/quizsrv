package main

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"sort"
)

func sitemap(w http.ResponseWriter, r *http.Request) {
	type entry struct {
		Loc string `xml:"loc"`
	}
	document := struct {
		XMLName xml.Name `xml:"urlset"`
		XMLNS   string   `xml:"xmlns,attr"`
		URLs    []entry  `xml:"url"`
	}{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: []entry{{Loc: "https://sne.dreadl0ck.net/"}}}
	var names []string
	for name := range data.Courses {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		document.URLs = append(document.URLs, entry{Loc: "https://sne.dreadl0ck.net/courses/" + url.PathEscape(name)})
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_ = xml.NewEncoder(w).Encode(document)
}
