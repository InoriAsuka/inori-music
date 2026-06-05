# syntax=docker/dockerfile:1

FROM golang:1.25.1-alpine AS build
WORKDIR /src
COPY go.work ./
COPY services/api/go.mod ./services/api/go.mod
RUN go work sync
COPY services/api ./services/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/inori-api ./services/api/cmd/server

FROM alpine:3.22
RUN addgroup -S inori && adduser -S -G inori inori && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/inori-api /usr/local/bin/inori-api
RUN mkdir -p /data && chown -R inori:inori /data
USER inori
ENV INORI_HTTP_ADDR=0.0.0.0:8080 \
    INORI_STORAGE_REPOSITORY_FILE=/data/storage-backends.json \
    INORI_MEDIA_OBJECT_REPOSITORY_FILE=/data/media-objects.json
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/inori-api"]
