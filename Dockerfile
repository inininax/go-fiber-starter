# --- 빌드 스테이지 ---
FROM golang:1.27-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG COMMIT=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
	-ldflags="-s -w -X main.commit=${COMMIT}" \
	-o /out/api ./cmd/api && \
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# --- 실행 스테이지 ---
FROM alpine:3.20
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /out/migrate /app/migrate
USER app
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- http://127.0.0.1:8080/livez || exit 1

ENTRYPOINT ["/app/api"]
