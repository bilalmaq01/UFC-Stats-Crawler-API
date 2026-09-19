# syntax=docker/dockerfile:1

# ---- build ----
FROM golang:1.26 AS build
WORKDIR /src

# Cache modules first.
COPY go.mod go.sum ./
RUN go mod download

# Build a fully static binary (no cgo) so it runs on distroless/static.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o /crawler ./cmd/crawler

# ---- run ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /crawler /crawler
ENTRYPOINT ["/crawler"]
# Default mode; the ECS task definition's `command` overrides this.
CMD ["all", "5"]
