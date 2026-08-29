FROM scratch

# timezone setup
ENV ZONEINFO /zoneinfo.zip
ADD static /zoneinfo.zip

COPY mdtopdf /usr/bin/mdtopdf
COPY quizsrv /usr/sbin/quizsrv
COPY files /etc/quizsrv/files
COPY pages /etc/quizsrv/pages
COPY courses /etc/quizsrv/courses

# Run as a non-root user on an unprivileged port.
#
# Deployed as a Dokploy APPLICATION (a swarm service). Dokploy populates only 8
# ContainerSpec fields, none security-related: no user, no capability drop, no
# PID limit, and `docker service update` is reverted by the next deploy
# (server/todos/dokploy-swarm-no-security-fields.md). The image is the only
# place this can be fixed, and USER is the only hardening it can receive.
#
# Running non-root empties the effective capability set — from root with
# Docker's default 14 capabilities to none.
#
# THE PORT MOVE IS PART OF THE SAME CHANGE. A non-root process cannot bind :80
# without CAP_NET_BIND_SERVICE, which swarm cannot grant here, so the listener
# moves to :8080 and the Dokploy domain's port must move with it. Out of step,
# the site 502s.
#
# Numeric uid because this is a `scratch` image: there is no /etc/passwd for a
# name to resolve against. Everything the server reads under /etc/quizsrv is
# COPYed world-readable, so no chown is required.
USER 65532:65532

EXPOSE 8080

ARG VERSION
ENV cacheBUSTER VERSION

# start quizsrv
ENTRYPOINT ["/usr/sbin/quizsrv", "-c=/etc/quizsrv", "-addr=:8080", "-tls=false"]
