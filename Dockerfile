FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta \
    UPLOAD_DIR=/data/uploads
WORKDIR /app
COPY --from=build /out/server /app/server
COPY seed-uploads /app/seed-uploads
EXPOSE 8000
# Salin foto menu awal ke volume (tidak menimpa file yang sudah ada), lalu jalankan server.
CMD ["sh", "-c", "mkdir -p \"$UPLOAD_DIR/menu\" && cp -rn /app/seed-uploads/. \"$UPLOAD_DIR/\" || true; exec /app/server"]
