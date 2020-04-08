FROM lambci/lambda:build-go1.x

WORKDIR /go/src/github.com/briankopp/hasta-la-vista
COPY go.* ./
RUN go mod download
COPY . .
RUN cd cmd/lambda-spot-termination && \
    go build && \
    zip lambda.zip lambda-spot-termination
RUN ls -al cmd/lambda-spot-termination
