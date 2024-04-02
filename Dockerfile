# Build Stage
FROM golang:1.22.0-alpine3.19 AS build
WORKDIR /streamfair_user_svc
COPY . .

# Install git
RUN apk update && apk add --no-cache git

# Set GOPRIVATE environment variable
ENV GOPRIVATE=github.com/Streamfair/streamfair_user_svc,github.com/Streamfair/streamfair_session_svc,github.com/Streamfair/streamfair_token_svc,github.com/Streamfair/streamfair_idp

# Add .netrc file for GitHub authentication
COPY .netrc /root/.netrc
RUN chmod 600 /root/.netrc

RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o user_svc main.go

# Run Stage
FROM alpine:3.19
WORKDIR /streamfair_user_svc

# Copy the binary from the build stage
COPY --from=build /streamfair_user_svc/user_svc .

COPY sh ./sh
COPY db/migration ./db/migration

EXPOSE 8084
EXPOSE 9094

CMD [ "/streamfair_user_svc/user_svc" ]
ENTRYPOINT [ "/streamfair_user_svc/start.sh" ]

# Install bash, curl, and git in the final image
RUN apk add --no-cache bash curl git
