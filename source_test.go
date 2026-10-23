package main

import (
	"crypto/sha256"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceAccessAndIntegrity(t *testing.T) {
	dir := t.TempDir()
	oldDir, oldRevision, oldHash := *configFolder, SourceRevision, SourceSHA256
	t.Cleanup(func() { *configFolder, SourceRevision, SourceSHA256 = oldDir, oldRevision, oldHash })
	*configFolder = dir
	SourceRevision = strings.Repeat("a", 40)
	archive := []byte("synthetic source archive")
	SourceSHA256 = fmt.Sprintf("%x", sha256.Sum256(archive))
	if err := os.Mkdir(filepath.Join(dir, "source"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "source", SourceRevision+".tar.gz")
	if err := os.WriteFile(path, archive, 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/source", "/api/source", "/source/archive"} {
		r := httptest.NewRecorder()
		sourceHandler(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 200 || r.Header().Get("X-Source-SHA256") != SourceSHA256 {
			t.Fatalf("%s: %d", path, r.Code)
		}
		if path == "/source" && !strings.Contains(r.Body.String(), `href="/source/archive"`) {
			t.Fatal("missing visible download")
		}
		if path == "/source/archive" && r.Body.String() != string(archive) {
			t.Fatal("wrong download bytes")
		}
	}
	if err := os.WriteFile(path, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRecorder()
	sourceHandler(r, httptest.NewRequest("GET", "/source/archive", nil))
	if r.Code != 503 {
		t.Fatal("accepted a mismatched archive")
	}
	SourceRevision = "../../secret"
	r = httptest.NewRecorder()
	sourceHandler(r, httptest.NewRequest("GET", "/api/source", nil))
	if r.Code != 503 {
		t.Fatal("accepted traversal revision")
	}
}

func TestFontAndTemplateDelivery(t *testing.T) {
	font, err := os.ReadFile("files/Inter.woff2")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(font)) != "2dcd3612039a9aaa2660c0c9edd0f4b982e19d317af5ffe4161c37ddc803b5d0" {
		t.Fatal("font changed")
	}
	license, err := os.ReadFile("files/Inter-OFL.txt")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(license)) != "5b9321a4298cfeb6b34354164a1c3afc3db114569984c502b9b35d988fd58c57" {
		t.Fatal("license changed")
	}
	for _, name := range []string{"index", "course", "quiz", "pdf"} {
		b, err := os.ReadFile("pages/" + name + ".html")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "SFPro") || !strings.Contains(string(b), "/files/Inter.woff2") || !strings.Contains(string(b), `href="/source"`) {
			t.Fatalf("%s missing replacement/source link", name)
		}
	}
}
