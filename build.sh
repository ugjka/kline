#!/bin/bash
go build
rsync -Pvhr --times . server.internal:kline/
