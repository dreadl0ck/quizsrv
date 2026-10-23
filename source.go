package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var SourceRevision = "development"
var SourceSHA256 = "unpublished"
var sourceRevisionPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var sourceHashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var sourcePage = template.Must(template.New("source").Parse(`<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Matching server source</title><main style="max-width:52rem;margin:3rem auto;padding:1rem;font:18px/1.6 system-ui;overflow-wrap:anywhere"><h1>Matching server source</h1><p>This modified quiz server combines GPLv3 application code with the AGPLv3-or-later Anki library. All remote users may obtain its full corresponding source, covered dependencies and build/install instructions without an account or charge.</p><p><a href="/source/archive">Download matching source</a> · <a href="/api/source">Source metadata</a> · <a href="/files/Inter-OFL.txt">Inter font license</a></p><p>Revision: {{.Revision}}</p><p>Archive SHA-256: {{.SHA256}}</p><p>Original component terms are retained; the combined work's network-source obligation includes its GPLv3 parts. Source delivery does not license course PDFs, Anki decks or other third-party teaching material. Production data and credentials are excluded.</p></main></html>`))

func sourceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/source" && r.URL.Path != "/source/archive" && r.URL.Path != "/api/source" {
		http.NotFound(w, r)
		return
	}
	if !sourceRevisionPattern.MatchString(SourceRevision) || !sourceHashPattern.MatchString(SourceSHA256) {
		http.Error(w, "Matching source has not been published", http.StatusServiceUnavailable)
		return
	}
	archive, err := os.ReadFile(filepath.Join(*configFolder, "source", SourceRevision+".tar.gz"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(archive)) != SourceSHA256 {
		http.Error(w, "Matching source is unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Source-Revision", SourceRevision)
	w.Header().Set("X-Source-SHA256", SourceSHA256)
	if r.URL.Path == "/source/archive" {
		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", `attachment; filename="source-`+SourceRevision+`.tar.gz"`)
		if r.Method != http.MethodHead {
			w.Write(archive)
		}
		return
	}
	if r.URL.Path == "/api/source" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"revision": SourceRevision, "sha256": SourceSHA256, "archive": "/source/archive", "license": "GPL-3.0-only AND AGPL-3.0-or-later"})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	sourcePage.Execute(w, struct{ Revision, SHA256 string }{SourceRevision, SourceSHA256})
}
