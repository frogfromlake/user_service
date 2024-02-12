# Build Stage
FROM golang:1.22.0-alpine3.19 AS build
WORKDIR /streamfair_user_svc
COPY . .
RUN go build -o user_svc main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz

# Run Stage
FROM alpine:3.19
WORKDIR /streamfair_user_svc

# Copy the binary from the build stage
COPY --from=build /streamfair_user_svc/user_svc .
# Copy the downloaded migration binary from the build stage
COPY --from=build /streamfair_user_svc/migrate ./migrate

COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./migration

EXPOSE 8084
EXPOSE 9094

CMD [ "/streamfair_user_svc/user_svc" ]
ENTRYPOINT [ "/streamfair_user_svc/start.sh" ]