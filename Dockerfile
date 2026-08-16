# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o reversedrop ./cmd/reversedrop
RUN go build -o reversedrop-gui ./gui

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache \
    libc6-compat \
    libx11 \
    libxext \
    libxrender \
    libxrandr \
    libxcursor \
    libxi \
    libxtst \
    libxinerama \
    libwayland \
    libxkbcommon \
    libdrm \
    libgbm \
    mesa-gl \
    mesa-dri-gallium

WORKDIR /app

COPY --from=builder /app/reversedrop .
COPY --from=builder /app/reversedrop-gui .

RUN mkdir -p /app/downloads

EXPOSE 8770

CMD ["./reversedrop-gui"]
