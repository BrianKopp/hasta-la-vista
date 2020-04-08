FROM lambci/lambda:build-go1.x

WORKDIR /go/src/github.com/briankopp/hasta-la-vista
COPY go.* ./
RUN go mod download
COPY . .
RUN cd cmd/lambda-asg-term-hook && \
    go build && \
    zip lambda.zip lambda-asg-term-hook
RUN ls -al cmd/lambda-asg-term-hook
