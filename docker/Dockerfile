FROM docker.io/golang:1.12.4-stretch

COPY . /proxy
WORKDIR /proxy

RUN go build

EXPOSE 8888/tcp 5683/udp

COPY ./docker/start.sh /start.sh

ENTRYPOINT ["/start.sh"]
