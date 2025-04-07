FROM golang:1.17.3-alpine3.14

RUN mkdir /app

WORKDIR /app

COPY . .

RUN go build main.go

ENTRYPOINT ["./main"]