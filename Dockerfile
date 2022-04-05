FROM golang:1.18.0-alpine as builder

ARG appVersion
ARG gitUrl="git@github.com:"
ARG token="9f77928e6f1ff8b58ad51df6b60f8fe75cdf3c1d"

RUN apk add --no-cache --update 'git' 'build-base' && \
    git config --global url."https://9f77928e6f1ff8b58ad51df6b60f8fe75cdf3c1d:@github.com/".insteadOf "https://github.com/"


ADD . /content-marketing
WORKDIR /content-marketing

RUN go install

ENV PORT=8080 \
    GIN_MODE=release \
    ENV=production \
    APPNAME="content-marketing" \
    APPVERSION="initial" \
    ENV=production


ENTRYPOINT ["/go/bin/content-marketing"]
EXPOSE 8080
