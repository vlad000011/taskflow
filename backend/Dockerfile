FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/taskflow ./cmd/api
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -H app
WORKDIR /app
COPY --from=build /out/taskflow /app/taskflow
COPY migrations /app/migrations
COPY docs /app/docs
RUN mkdir -p /app/uploads && chown -R app:app /app
USER app
EXPOSE 8081
CMD ["/app/taskflow"]
