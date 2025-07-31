# BUILD: Compile the Go binary
FROM golang:1.24.4-alpine AS build
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o kvs ./cmd/server

# Stage 2: Build the Key-Value Store image
FROM scratch 
# Copy the binary from the build container 
COPY --from=build /app/kvs / 
EXPOSE 8080 
CMD ["/kvs"]