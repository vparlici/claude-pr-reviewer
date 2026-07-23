# ---- Build stage ----
FROM golang:1.22-alpine AS build

WORKDIR /src

# Cache dependencies first for faster incremental builds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Statically linked, stripped binary for a tiny final image.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/reviewer .

# ---- Runtime stage ----
# distroless/static ships CA certificates (needed for HTTPS to the GitHub and Anthropic APIs) and nothing else — no shell, minimal attack surface.
FROM gcr.io/distroless/static:nonroot

COPY --from=build /out/reviewer /reviewer

ENTRYPOINT ["/reviewer"]