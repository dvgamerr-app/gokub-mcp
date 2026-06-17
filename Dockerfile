FROM golang:1.25.5-alpine AS build
ENV GOTOOLCHAIN=local
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /gokub-mcp .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
COPY --from=build /gokub-mcp /usr/local/bin/gokub-mcp

USER app
WORKDIR /home/app

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["gokub-mcp", "-serv"]
