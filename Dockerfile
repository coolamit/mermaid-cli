FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o mmd-cli ./cmd/mmd-cli

FROM alpine:3.21

# Install Chromium
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont \
    font-noto-emoji

# Tell chromedp where to find the browser
ENV CHROME_BIN=/usr/bin/chromium-browser

WORKDIR /data

COPY --from=builder /app/mmd-cli /usr/local/bin/mmd-cli

ENTRYPOINT ["mmd-cli"]
