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

ENTRYPOINT ["/my-app"]