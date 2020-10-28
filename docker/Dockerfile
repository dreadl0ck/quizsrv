FROM scratch

# timezone setup
ENV ZONEINFO /zoneinfo.zip
ADD static /zoneinfo.zip

COPY mdtopdf /usr/bin/mdtopdf
COPY quizsrv /usr/sbin/quizsrv
COPY files /etc/quizsrv/files
COPY pages /etc/quizsrv/pages
COPY courses /etc/quizsrv/courses

EXPOSE 443
EXPOSE 80

# start quizsrv
ENTRYPOINT ["/usr/sbin/quizsrv", "-c=/etc/quizsrv", "-addr=:443", "-tls"]
