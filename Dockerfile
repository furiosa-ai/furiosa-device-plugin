FROM golang:1.27.2-bookworm AS builder

WORKDIR /workspace
COPY . .

# Build device-plugin binary
RUN CGO_ENABLED=0 go build -o /main cmd/main.go

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /

COPY --from=builder /main /main

CMD ["./main"]
