FROM golang:latest

RUN apt-get update && apt-get upgrade

LABEL maintainer="Sorova Anna"

WORKDIR /app

COPY go.mod ./

RUN go mod download         # Dependencies will be cached if the go.mod file are not changed

COPY . .                    # Copy the source from the current directory to the Working Directory inside the container

RUN go build -o server .

EXPOSE 9000

CMD ["./server"]
