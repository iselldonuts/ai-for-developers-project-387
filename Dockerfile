FROM node:24-alpine AS frontend

WORKDIR /src/web

COPY web/package*.json ./
RUN npm ci

COPY web ./
ENV VITE_API_BASE_URL=
RUN npm run build

FROM golang:1.26.1-alpine AS backend

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/calcal ./cmd/calcal

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache tzdata \
    && addgroup -S calcal \
    && adduser -S calcal -G calcal

COPY --from=backend /out/calcal ./calcal
COPY --from=backend /src/migrations ./migrations
COPY --from=frontend /src/web/dist ./web/dist

USER calcal

EXPOSE 8080

CMD ["./calcal"]
