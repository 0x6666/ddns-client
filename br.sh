# !/bin/bash

set -e

if [ -a "$GOPATH/bin/ddns-client" ]; then
	rm "$GOPATH/bin/ddns-client"
fi

go install github.com/inimei/ddns-client

cfg=$GOPATH/bin/ddns-client.toml
if [ -a "$cfg" ]; then
	echo "$cfg already exist...."
else 
	ln -s $GOPATH/src/github.com/inimei/ddns-client/ddns-client.toml $cfg
fi

$GOPATH/bin/ddns-client
