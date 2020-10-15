#!/bin/bash
#
# ZEUS Error Dump
# Timestamp: [Thu Oct 15 16:28:13 2020]
# Error: exit status 1
# StdErr: 
# ../../go-acme/lego/providers/dns/vegadns/vegadns.go:10:2: cannot find package "github.com/OpenDNS/vegadns2client" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/OpenDNS/vegadns2client (from $GOROOT)
# 	/Users/alien/go/src/github.com/OpenDNS/vegadns2client (from $GOPATH)
# ../../go-acme/lego/providers/dns/fastdns/fastdns.go:10:2: cannot find package "github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v1" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v1 (from $GOROOT)
# 	/Users/alien/go/src/github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v1 (from $GOPATH)
# ../../go-acme/lego/providers/dns/fastdns/fastdns.go:11:2: cannot find package "github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid (from $GOROOT)
# 	/Users/alien/go/src/github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid (from $GOPATH)
# ../../go-acme/lego/providers/dns/dnsimple/dnsimple.go:12:2: cannot find package "github.com/dnsimple/dnsimple-go/dnsimple" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/dnsimple/dnsimple-go/dnsimple (from $GOROOT)
# 	/Users/alien/go/src/github.com/dnsimple/dnsimple-go/dnsimple (from $GOPATH)
# ../../go-acme/lego/providers/dns/exoscale/exoscale.go:11:2: cannot find package "github.com/exoscale/egoscale" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/exoscale/egoscale (from $GOROOT)
# 	/Users/alien/go/src/github.com/exoscale/egoscale (from $GOPATH)
# ../../go-acme/lego/providers/dns/designate/designate.go:14:2: cannot find package "github.com/gophercloud/gophercloud" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/gophercloud/gophercloud (from $GOROOT)
# 	/Users/alien/go/src/github.com/gophercloud/gophercloud (from $GOPATH)
# ../../go-acme/lego/providers/dns/designate/designate.go:15:2: cannot find package "github.com/gophercloud/gophercloud/openstack" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/gophercloud/gophercloud/openstack (from $GOROOT)
# 	/Users/alien/go/src/github.com/gophercloud/gophercloud/openstack (from $GOPATH)
# ../../go-acme/lego/providers/dns/designate/designate.go:16:2: cannot find package "github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets (from $GOROOT)
# 	/Users/alien/go/src/github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets (from $GOPATH)
# ../../go-acme/lego/providers/dns/designate/designate.go:17:2: cannot find package "github.com/gophercloud/gophercloud/openstack/dns/v2/zones" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/gophercloud/gophercloud/openstack/dns/v2/zones (from $GOROOT)
# 	/Users/alien/go/src/github.com/gophercloud/gophercloud/openstack/dns/v2/zones (from $GOPATH)
# ../../go-acme/lego/providers/dns/iij/iij.go:12:2: cannot find package "github.com/iij/doapi" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/iij/doapi (from $GOROOT)
# 	/Users/alien/go/src/github.com/iij/doapi (from $GOPATH)
# ../../go-acme/lego/providers/dns/iij/iij.go:13:2: cannot find package "github.com/iij/doapi/protocol" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/iij/doapi/protocol (from $GOROOT)
# 	/Users/alien/go/src/github.com/iij/doapi/protocol (from $GOPATH)
# ../../go-acme/lego/providers/dns/linodev4/linodev4.go:15:2: cannot find package "github.com/linode/linodego" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/linode/linodego (from $GOROOT)
# 	/Users/alien/go/src/github.com/linode/linodego (from $GOPATH)
# ../../go-acme/lego/providers/dns/liquidweb/liquidweb.go:13:2: cannot find package "github.com/liquidweb/liquidweb-go/client" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/liquidweb/liquidweb-go/client (from $GOROOT)
# 	/Users/alien/go/src/github.com/liquidweb/liquidweb-go/client (from $GOPATH)
# ../../go-acme/lego/providers/dns/liquidweb/liquidweb.go:14:2: cannot find package "github.com/liquidweb/liquidweb-go/network" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/liquidweb/liquidweb-go/network (from $GOROOT)
# 	/Users/alien/go/src/github.com/liquidweb/liquidweb-go/network (from $GOPATH)
# ../../go-acme/lego/providers/dns/namedotcom/namedotcom.go:13:2: cannot find package "github.com/namedotcom/go/namecom" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/namedotcom/go/namecom (from $GOROOT)
# 	/Users/alien/go/src/github.com/namedotcom/go/namecom (from $GOPATH)
# ../../go-acme/lego/providers/dns/inwx/inwx.go:12:2: cannot find package "github.com/nrdcg/goinwx" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/nrdcg/goinwx (from $GOROOT)
# 	/Users/alien/go/src/github.com/nrdcg/goinwx (from $GOPATH)
# ../../go-acme/lego/providers/dns/namesilo/namesilo.go:12:2: cannot find package "github.com/nrdcg/namesilo" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/nrdcg/namesilo (from $GOROOT)
# 	/Users/alien/go/src/github.com/nrdcg/namesilo (from $GOPATH)
# ../../go-acme/lego/providers/dns/oraclecloud/configprovider.go:11:2: cannot find package "github.com/oracle/oci-go-sdk/common" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/oracle/oci-go-sdk/common (from $GOROOT)
# 	/Users/alien/go/src/github.com/oracle/oci-go-sdk/common (from $GOPATH)
# ../../go-acme/lego/providers/dns/oraclecloud/oraclecloud.go:13:2: cannot find package "github.com/oracle/oci-go-sdk/dns" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/oracle/oci-go-sdk/dns (from $GOROOT)
# 	/Users/alien/go/src/github.com/oracle/oci-go-sdk/dns (from $GOPATH)
# ../../go-acme/lego/providers/dns/ovh/ovh.go:14:2: cannot find package "github.com/ovh/go-ovh/ovh" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/ovh/go-ovh/ovh (from $GOROOT)
# 	/Users/alien/go/src/github.com/ovh/go-ovh/ovh (from $GOPATH)
# ../../go-acme/lego/providers/dns/sakuracloud/client.go:9:2: cannot find package "github.com/sacloud/libsacloud/api" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/sacloud/libsacloud/api (from $GOROOT)
# 	/Users/alien/go/src/github.com/sacloud/libsacloud/api (from $GOPATH)
# ../../go-acme/lego/providers/dns/sakuracloud/client.go:10:2: cannot find package "github.com/sacloud/libsacloud/sacloud" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/sacloud/libsacloud/sacloud (from $GOROOT)
# 	/Users/alien/go/src/github.com/sacloud/libsacloud/sacloud (from $GOPATH)
# ../../go-acme/lego/providers/dns/linode/linode.go:12:2: cannot find package "github.com/timewasted/linode/dns" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/timewasted/linode/dns (from $GOROOT)
# 	/Users/alien/go/src/github.com/timewasted/linode/dns (from $GOPATH)
# ../../go-acme/lego/providers/dns/transip/transip.go:13:2: cannot find package "github.com/transip/gotransip" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/transip/gotransip (from $GOROOT)
# 	/Users/alien/go/src/github.com/transip/gotransip (from $GOPATH)
# ../../go-acme/lego/providers/dns/transip/transip.go:14:2: cannot find package "github.com/transip/gotransip/domain" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/transip/gotransip/domain (from $GOROOT)
# 	/Users/alien/go/src/github.com/transip/gotransip/domain (from $GOPATH)
# ../../go-acme/lego/providers/dns/vultr/vultr.go:17:2: cannot find package "github.com/vultr/govultr" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/github.com/vultr/govultr (from $GOROOT)
# 	/Users/alien/go/src/github.com/vultr/govultr (from $GOPATH)
# ../../../google.golang.org/api/internal/pool.go:10:2: cannot find package "google.golang.org/grpc/naming" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/google.golang.org/grpc/naming (from $GOROOT)
# 	/Users/alien/go/src/google.golang.org/grpc/naming (from $GOPATH)
# ../../go-acme/lego/providers/dns/ns1/ns1.go:14:2: cannot find package "gopkg.in/ns1/ns1-go.v2/rest" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/gopkg.in/ns1/ns1-go.v2/rest (from $GOROOT)
# 	/Users/alien/go/src/gopkg.in/ns1/ns1-go.v2/rest (from $GOPATH)
# ../../go-acme/lego/providers/dns/ns1/ns1.go:15:2: cannot find package "gopkg.in/ns1/ns1-go.v2/rest/model/dns" in any of:
# 	/usr/local/Cellar/go/1.15.2/libexec/src/gopkg.in/ns1/ns1-go.v2/rest/model/dns (from $GOROOT)
# 	/Users/alien/go/src/gopkg.in/ns1/ns1-go.v2/rest/model/dns (from $GOPATH)
# 


#!/bin/bash
VERSION="v0.1"



cp -rf categories files pages docker
GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags "-X main.Version=${VERSION}" -o docker/quizsrv server.go
docker build -t dreadl0ck/quizsrv:${VERSION} docker
rm -vf docker/quizsrv
