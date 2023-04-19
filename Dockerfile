FROM golang:latest

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

COPY .env .env

RUN go build -o main .


CMD ["./streamlining-backend"]
