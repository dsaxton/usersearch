#!/usr/bin/env bash
go build -ldflags="-s -w" main.go
mv main usersearch