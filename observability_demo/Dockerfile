FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --silent
COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -o /encroach-demo ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=backend /encroach-demo /usr/local/bin/encroach-demo
COPY --from=backend /app/frontend/dist /app/frontend/dist
COPY --from=backend /app/db /app/db
WORKDIR /app
EXPOSE 3000
ENTRYPOINT ["encroach-demo"]
