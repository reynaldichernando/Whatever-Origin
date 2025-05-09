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

## Self-Hosting

The easiest way to self-host Whatever Origin is by using Docker.

### Prerequisites

- A Virtual Private Server (VPS) or any machine.
- Docker installed. You can find installation instructions on the [official Docker website](https://docs.docker.com/engine/install/).

### Steps

1.  Clone this repository

    ```bash
    git clone https://github.com/reynaldichernando/whatever-origin.git
    cd whatever-origin
    ```

2.  Run with Docker Compose

    This will pull the latest image and start the service.

    ```bash
    docker compose up -d
    ```

3.  Access the service

    You should be able to access your self-hosted Whatever Origin service on port 80:

    ```
    http://localhost:80
    ```

## Background

When I was facing Same Origin problems while developing [Bitcoin Pie](http://bitcoinpie.com/), I was excited to discover how anyorigin.com solved the issue for me ... only, a week later it stopped working for some https sites.

For example, right now try and feed https://bitcointalk.org/ into anyorigin and you'll get an ugly "null" as the output.

Having recently discovered Heroku and Play!, I found that deploying a simple server app is no longer a big deal, and so made out to develop a simple, open source alternative to Any Origin.
