FROM golang:1-26-alpine AS builder

WORKDIR /app

# copy source code
COPY main.go /app/
COPY config.go /app/
COPY server.go /app/
COPY logger.go /app/

# copy assets
COPY static /app/static
COPY templates /app/templates

# build the application
RUN go build -o cv-app .

FROM gcr.io/distroless/base-debian13

WORKDIR /app

# copy the built application from the builder stage
COPY --from=builder /app/cv-app /app/cv-app
COPY --from=builder /app/static /app/static
COPY --from=builder /app/templates /app/templates

EXPOSE 3000

CMD ["/app/cv-app"]