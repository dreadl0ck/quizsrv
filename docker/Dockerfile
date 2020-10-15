FROM scratch

# timezone setup
ENV ZONEINFO /zoneinfo.zip
ADD static /zoneinfo.zip

COPY quizsrv /usr/sbin/quizsrv
COPY files /etc/quizsrv/files
COPY pages /etc/quizsrv/pages
COPY categories /etc/quizsrv/categories

EXPOSE 443
EXPOSE 80

# start quizsrv
ENTRYPOINT ["/usr/sbin/quizsrv", "-c=/etc/quizsrv", "-addr=:443", "-tls"]
