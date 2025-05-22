FROM golang:1.24-alpine AS build

WORKDIR /app

COPY ../go.mod go.sum ./

RUN go mod download

COPY ../main.go .

RUN go build -o my-app .

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /app/my-app /my-app

EXPOSE 8888

USER nonroot

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/my-app"]