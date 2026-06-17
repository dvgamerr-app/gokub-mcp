# syntax=docker/dockerfile:1

# ---- build ----
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/bitkub-mcp .

# ---- run ----
FROM alpine:3.21
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
COPY --from=build /bin/bitkub-mcp /usr/local/bin/bitkub-mcp
USER app
ENV PORT=8080
EXPOSE 8080
# HTTP/SSE mode — stdio mode is only useful when a client spawns the binary directly.
ENTRYPOINT ["bitkub-mcp", "-serv"]
