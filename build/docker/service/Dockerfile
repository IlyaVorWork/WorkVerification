# ---------- stage 1 : build go ----------
FROM golang:1.25.5-alpine AS builder

WORKDIR /workVerification

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

WORKDIR /workVerification/cmd/app
RUN CGO_ENABLED=0 go build -o /bin/app

# ---------- stage 2 : final image ----------
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    adb \
    ca-certificates \
    yara \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /bin/app /bin/app

CMD ["/bin/app"]
