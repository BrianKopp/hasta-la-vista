# Build stage
FROM golang:alpine as build
RUN apk update && \
    apk --no-cache add git && \
    go get github.com/aws/aws-sdk-go && \
    go get github.com/rs/zerolog && \
    go get github.com/rs/zerolog/log
WORKDIR /go/src/app
COPY cmd/ ./
RUN go install .

# Deploy stage
FROM alpine
EXPOSE 3000
COPY --from=build /go/bin/app /app
ENTRYPOINT ["/app"]
