FROM golang:1.26.4-alpine AS build

WORKDIR /app

COPY go.mod go.sum main.go ./

RUN go mod download && \
    go build -o my-app .

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /app/my-app /my-app

EXPOSE 8888

USER nonroot

ENTRYPOINT ["/my-app"]
