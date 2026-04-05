# Build stage
FROM golang:1.26-alpine AS build

WORKDIR /app

# Copy go.mod (and go.sum if it exists)
COPY go.mod ./
COPY go.sum* ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

# Final stage — pin to a specific Alpine minor version to avoid surprise breakage.
# Bump intentionally when you need a newer Alpine (e.g., security patch).
FROM alpine:3.21

WORKDIR /

# Copy the binary from the build stage
COPY --from=build /server /server

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["/server"]
