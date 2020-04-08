# /bin/bash -e

docker build ../.. -f ./Dockerfile -t lambda-asg-term-hook:latest --no-cache

docker run --rm -it -v $(pwd)/out:/out \
    lambda-asg-term-hook:latest \
    cp cmd/lambda-asg-term-hook/lambda.zip /out/lambda.zip
