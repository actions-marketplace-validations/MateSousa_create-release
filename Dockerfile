FROM golang:1.20 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o create-release .

FROM gcr.io/distroless/static

COPY --from=builder /app/create-release /usr/local/bin/create-release

ENTRYPOINT ["/usr/local/bin/create-release"]


