#!/bin/bash
go build
rsync -Pvhr --times . server.internal:kline/
patch main.go < diff.patch
git reset --hard
