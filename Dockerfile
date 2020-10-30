FROM golang:1.15.3-alpine

WORKDIR /app

COPY go.mod go.sum main.go ./

RUN go build -o main

CMD ./main
