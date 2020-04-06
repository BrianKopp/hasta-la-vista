# /bin/bash -e

docker build ../.. -f ./Dockerfile -t lambda-spot-termination-build:latest --no-cache

docker run --rm -it -v $(pwd)/out:/out \
    lambda-spot-termination-build:latest \
    cp cmd/lambda-spot-termination/lambda.zip /out/lambda.zip
