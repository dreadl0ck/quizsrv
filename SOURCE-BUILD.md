# Matching source

First-party server code retains GPLv3. `github.com/flimzy/anki` at
`03460c56b85c` is AGPLv3-or-later. The combined executable's corresponding
source includes its GPLv3 parts under AGPLv3 §13; component terms are preserved,
not replaced by a blanket new grant. `vendor/` contains the covered dependency
code and original legal documents. The source archive excludes teaching
datasets, course PDFs, Anki decks and production records; no rights in those
materials are asserted by this source offer.

## Build

Use Go 1.25 or newer. The source archive supports offline compilation:

```sh
GOPROXY=off GOSUMDB=off go build -mod=vendor -o quizsrv .
GOPROXY=off GOSUMDB=off go build -mod=vendor -o mdtopdf ./cmd/mdtopdf
```

`release.py` builds Linux amd64 server/helper and assembles a fresh scratch
image context. `docker/Dockerfile` defines installation paths, non-root UID
65532 and startup flags (`-c=/etc/quizsrv -addr=:8080 -tls=false`). Provide your
own licensed course YAML at `/etc/quizsrv/courses`; an empty directory permits
an empty catalog. Templates and reviewed browser/font assets are included.
No service credentials are needed for the HTTP server build or startup.
SourceRevision/SourceSHA256 are linker values bound to the published archive.
Standalone binaries above build directly from the archive without Git/network.
Release scripts expect a Git checkout of the revision being packaged.

`zeus build-docker` requires `SOURCE_MANIFEST`, a fresh `RELEASE_OUTPUT` and an
explicit `RELEASE_IMAGE`. `zeus push-docker` also requires `SOURCE_PUBLIC_URL`;
`publish-release.py` checks its anonymous HTTPS hash and the image's embedded
archive before pushing. Add `--check` to the Python publisher to verify without
pushing. Historical exports must use that revision's committed packager.
No CI workflow was present or added.

The former opaque mdtopdf binary has SHA-256
`ddbfa9295c495cc037ea9fada7c55a8c471a3e280a17d915f55eda945fdf680e`.
It contains mandolyte/mdtopdf and jung-kurt/gofpdf symbols but no exact dependency
version receipt. New builds use the committed wrapper and pinned
`mandolyte/mdtopdf v1.0.0` with its original MIT terms and vendored dependencies.
This replacement does not establish historical binary redistribution rights.

Inter 4.001 is unchanged, SHA-256
`2dcd3612039a9aaa2660c0c9edd0f4b982e19d317af5ffe4161c37ddc803b5d0`.
`files/Inter-OFL.txt` and `Inter-provenance.json` retain full terms and original
embedded attribution. Third-party assets keep their own licenses.
