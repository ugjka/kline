#!/bin/bash
patch main.go < diff.patch
go build
rsync -Pvhr . server.internal:kline/
ssh server.internal -- sudo setcap 'cap_net_bind_service=ep' kline/kline
patch -R < diff.patch

# yeet
