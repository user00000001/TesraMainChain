# Build tesramain in a stock Go builder container
FROM golang:1.12-alpine as builder

RUN apk add --no-cache make gcc git musl-dev linux-headers

ADD . /TesraMainChain
RUN cd /TesraMainChain && make tesramain

# Pull Geth into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
RUN apk add --no-cache curl
COPY --from=builder /TesraMainChain/build/bin/tesramain /usr/local/bin/

EXPOSE 8545 17717/tcp 17717/udp
ENTRYPOINT ["tesramain"]
