FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS backend-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -o miebank .

FROM alpine:latest

WORKDIR /app
COPY --from=backend-builder /app/miebank ./
COPY --from=backend-builder /app/frontend/dist ./frontend/dist

EXPOSE 10000
CMD ["./miebank"]
