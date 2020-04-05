# Build stage
FROM golang:alpine as build
RUN apk update && \
    apk --no-cache add git
WORKDIR /go/src/github.com/briankopp/hasta-la-vista
COPY go.* ./
RUN go mod download
COPY . .
RUN go build
RUN ls -al

# Deploy stage
FROM alpine
EXPOSE 80
COPY --from=build /go/src/github.com/briankopp/hasta-la-vista/hasta-la-vista /app
ENTRYPOINT [ "/app" ]
