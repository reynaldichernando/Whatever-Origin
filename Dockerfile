FROM golang:1.24.3-alpine

RUN mkdir /app

WORKDIR /app

COPY . .

RUN go build main.go

ENTRYPOINT ["./main"]