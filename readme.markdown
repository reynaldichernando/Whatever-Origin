# Whatever Origin

Whatever Origin is an open source alternative to AnyOrigin.com

[Live site](http://whateverorigin.org/).

## Developing

Whatever Origin has been rewritten in Go.

To run the server locally, you need to have Go installed, you can get it from [here](https://go.dev/doc/install).

Clone the repository, and run the following commands:

```bash
go mod download
go run .
```

This will start the server on port 8080.

## Background

When I was facing Same Origin problems while developing [Bitcoin Pie](http://bitcoinpie.com/), I was excited to discover how anyorigin.com solved the issue for me ... only, a week later it stopped working for some https sites.

For example, right now try and feed https://bitcointalk.org/ into anyorigin and you'll get an ugly "null" as the output.

Having recently discovered Heroku and Play!, I found that deploying a simple server app is no longer a big deal, and so made out to develop a simple, open source alternative to Any Origin.
