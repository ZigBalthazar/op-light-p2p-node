FROM golang:1.22.2-alpine3.19 as builder

ENV REDIS_CONN_STRING="redis://localhost:6379"
ENV REDIS_STREAM_NAME="mantle"
ENV P2P_SEQUENCER_ADDRESS="0xAAC979CBeE00C75C35DE9a2635d8B75940F466dc" 

ENV GOPROXY=https://goproxy.io,direct

WORKDIR /app

COPY . .

RUN go build -o op-light-node .

EXPOSE 3000

CMD ["./op-light-node --rollup.config=./config/mantle/rollup.json"]