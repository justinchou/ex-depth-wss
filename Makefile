
all: build run

build:
	go build ./main.go

run:
	./main

tag:
  docker build -t ex-depth:1.0.5 .