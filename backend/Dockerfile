FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy mod files
COPY go.mod ./

# Copy source
COPY . .

# Tidy to generate go.sum and download deps
RUN go mod tidy

# Build
RUN go build -o /exchange-backend main.go


FROM alpine:latest
WORKDIR /app
COPY --from=builder /exchange-backend /app/exchange-backend

EXPOSE 8080
CMD ["/app/exchange-backend"]
