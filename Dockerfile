FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/Actual-Outcomes/doit/internal/version.Number=$(git describe --tags --always 2>/dev/null || echo dev)" -o /doit-server ./cmd/doit-server
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/Actual-Outcomes/doit/internal/version.Number=$(git describe --tags --always 2>/dev/null || echo dev)" -o /doit ./cmd/doit

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /doit-server /usr/local/bin/doit-server
COPY --from=builder /doit /usr/local/bin/doit
EXPOSE 8080
ENTRYPOINT ["doit-server"]
