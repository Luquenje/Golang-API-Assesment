@echo off

REM Get command
go get -d -v .\...

REM Build command
go build -o bin\program

REM Run command
go run .