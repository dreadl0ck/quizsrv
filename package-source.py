#!/usr/bin/env python3
"""Export committed code and vendored dependencies, never teaching/runtime data."""
import argparse
import gzip
import hashlib
import io
import json
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile

ROOT = Path(__file__).resolve().parent
FILES = {"LICENSE", "go.mod", "go.sum", "package-source.py", "release.py", "SOURCE-BUILD.md", "source-compose.yml", "sw.js", "manifest.json"}
ASSETS = {"Inter.woff2", "Inter-OFL.txt", "Inter-provenance.json", "Browser-LICENSES.txt", "nosleep.js", "swiped-events.js", "quiz.js", "fullscreen.svg", "gesture.svg", "keyboard.svg", "mouse.svg"}


def git(*args):
    return subprocess.check_output(["git", "-C", str(ROOT), *args])


def selected(name):
    p = Path(name)
    return name in FILES or (len(p.parts) == 1 and p.suffix == ".go") or p.parts[0] in {"cmd", "licenses"} or (p.parts[0] == "pages" and p.suffix == ".html") or (p.parts[0] == "files" and p.name in ASSETS) or name == "docker/Dockerfile"


def check(name, mode, data):
    if mode not in {"100644", "100755"} or any(p.startswith(".env") or p in {".git", "courses", "node_modules", "__pycache__"} for p in Path(name).parts):
        raise ValueError("Forbidden source member: " + name)
    if re.search(rb"-----BEGIN (?:[A-Z]+ )?PRIVATE KEY-----|(?:sk-ant-api|sk-ant-oat|ghp_)[A-Za-z0-9_-]{20,}", data):
        raise ValueError("Secret candidate: " + name)


def package(revision, output):
    revision = git("rev-parse", "--verify", revision + "^{commit}").decode().strip()
    members = []
    with tempfile.TemporaryDirectory() as temp:
        root = Path(temp)
        for entry in filter(None, git("ls-tree", "-rz", "--full-tree", revision).split(b"\0")):
            meta, name = entry.split(b"\t", 1)
            name = name.decode()
            if not selected(name): continue
            mode, kind, oid = meta.decode().split()
            data = git("cat-file", "blob", oid) if kind == "blob" else b""
            check(name, mode, data)
            path = root / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(data)
            members.append((name, mode, data))
        required = {"LICENSE", "go.mod", "go.sum", "source.go", "SOURCE-BUILD.md", "cmd/mdtopdf/main.go", "files/Inter-OFL.txt"}
        if not required <= {m[0] for m in members}: raise ValueError("Missing build source")
        subprocess.run(["go", "mod", "vendor"], cwd=root, check=True)
        for path in sorted((root / "vendor").rglob("*")):
            if not path.is_file(): continue
            if path.is_symlink(): raise ValueError("Vendored symlink")
            name, data = path.relative_to(root).as_posix(), path.read_bytes()
            check(name, "100644", data)
            members.append((name, "100644", data))
    output.mkdir(parents=True, exist_ok=True)
    archive = output / (revision + ".tar.gz")
    with archive.open("wb") as raw, gzip.GzipFile(filename="", mode="wb", fileobj=raw, mtime=0) as zipped:
        with tarfile.open(fileobj=zipped, mode="w", format=tarfile.PAX_FORMAT) as tar:
            for name, mode, data in sorted(members):
                item = tarfile.TarInfo("quiz-source-" + revision + "/" + name)
                item.size, item.mode, item.mtime = len(data), int(mode[-3:], 8), 0
                tar.addfile(item, io.BytesIO(data))
    manifest = {"revision": revision, "sha256": hashlib.sha256(archive.read_bytes()).hexdigest(), "archive": "/source/archive", "license": "GPL-3.0-only AND AGPL-3.0-or-later", "files": len(members), "scope": "Modified server and helper, templates, reviewed browser/font assets, full vendored build dependencies and original terms; no course datasets, PDF/Anki collections, secrets or runtime data."}
    (output / (revision + ".json")).write_text(json.dumps(manifest, indent=2) + "\n")
    print(json.dumps(manifest))


if __name__ == "__main__":
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("revision")
    p.add_argument("output", type=Path)
    a = p.parse_args()
    package(a.revision, a.output.resolve())
