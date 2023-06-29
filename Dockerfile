FROM redis:latest

EXPOSE 6379

CMD ["redis-server"]

FROM golang:1.19-alpine

WORKDIR /app

# Copy the Go module files
COPY go.mod .

# Download and install Go dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

EXPOSE 6600

CMD ["./main"]