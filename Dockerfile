# Build Stage
FROM golang:1.22.0-alpine3.19 AS build
WORKDIR /streamfair_user_service
COPY . .
RUN go build -o streamfair_user_service main.go

# Run Stage
FROM alpine:3.19
WORKDIR /streamfair_user_service
COPY --from=build /streamfair_user_service/streamfair_user_service .
COPY app.env .

EXPOSE 8084
EXPOSE 9094

CMD ["./streamfair_user_service"]