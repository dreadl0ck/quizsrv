# syntax=docker/dockerfile:1.7
FROM golang:1.27.4-alpine3.24@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS build
ARG SOURCE_REVISION
ARG SOURCE_SHA256
ARG VERSION=source
ADD --checksum=sha256:${SOURCE_SHA256} https://sne.dreadl0ck.net/legal/source/${SOURCE_REVISION}.tar.gz /source.tar.gz
WORKDIR /src
RUN tar -xzf /source.tar.gz --strip-components=1 && mkdir -p /payload
RUN --network=none CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.SourceRevision=${SOURCE_REVISION} -X main.SourceSHA256=${SOURCE_SHA256}" -o /payload/quizsrv . && CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags="-s -w" -o /payload/mdtopdf ./cmd/mdtopdf
COPY pages /payload/pages
COPY files /payload/files
COPY courses /payload/courses
RUN mkdir -p /payload/source /payload/licenses && cp /source.tar.gz /payload/source/${SOURCE_REVISION}.tar.gz && cp LICENSE SOURCE-BUILD.md sw.js manifest.json /payload/ && cp /usr/local/go/lib/time/zoneinfo.zip /payload/zoneinfo.zip && cp /usr/local/go/LICENSE /payload/licenses/Go-LICENSE.txt && find vendor -type f \( -iname 'license*' -o -iname 'copying*' -o -iname 'notice*' -o -iname 'authors*' \) -exec sh -c 'for file do mkdir -p "/payload/licenses/$(dirname "$file")"; cp "$file" "/payload/licenses/$file"; done' sh '{}' +

FROM scratch
ARG SOURCE_REVISION
ARG SOURCE_SHA256
LABEL org.opencontainers.image.revision=${SOURCE_REVISION} com.amsterdam-technologies.source.sha256=${SOURCE_SHA256}
ENV ZONEINFO=/zoneinfo.zip
COPY --from=build /payload/zoneinfo.zip /zoneinfo.zip
COPY --from=build /payload/mdtopdf /usr/bin/mdtopdf
COPY --from=build /payload/quizsrv /usr/sbin/quizsrv
COPY --from=build /payload/pages /etc/quizsrv/pages
COPY --from=build /payload/files /etc/quizsrv/files
COPY --from=build /payload/courses /etc/quizsrv/courses
COPY --from=build /payload/source /etc/quizsrv/source
COPY --from=build /payload/licenses /etc/quizsrv/licenses
COPY --from=build /payload/LICENSE /payload/SOURCE-BUILD.md /payload/sw.js /payload/manifest.json /etc/quizsrv/
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/usr/sbin/quizsrv", "-c=/etc/quizsrv", "-addr=:8080", "-tls=false"]
