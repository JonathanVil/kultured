# Stage 1 — build the Svelte frontend
FROM node:alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2 — build the Go binary
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o kultured .

# Stage 3 — minimal runtime image
FROM alpine
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /data
COPY --from=builder /app/kultured /usr/local/bin/kultured
EXPOSE 8085
ENV DB_PATH=/data/brew.db
CMD ["kultured"]
