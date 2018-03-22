#! /bin/env bash

dropdb jerver
createdb jerver
psql jerver -f ../schema/0.sql
go run initData.go
