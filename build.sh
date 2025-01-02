#!/bin/bash
patch main.go < diff.patch
go build
rsync -Pvhr --times . server.internal:kline/
git reset --hard

# yeet
