FROM golang:1.25.4-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY back ./back
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server ./back

FROM debian:bookworm-slim
WORKDIR /app/back
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN useradd --create-home --uid 10001 appuser && mkdir -p /app/uploads /app/seed_uploads
COPY --from=build /app/server /app/server
COPY front /app/front
COPY railway-entrypoint.sh /app/railway-entrypoint.sh
RUN chmod +x /app/railway-entrypoint.sh && chown -R appuser:appuser /app
ENV APP_ENV=production
ENV UPLOADS_DIR=/app/uploads
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/railway-entrypoint.sh"]
