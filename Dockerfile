# syntax=docker/dockerfile:1
# A sample recipe microservice in Go packaged into a container image.

FROM golang:latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-go-recipe
CMD ["/docker-go-recipe"]
