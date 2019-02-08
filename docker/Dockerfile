FROM docker.io/golang:1.11.2-stretch

COPY . /proxy
WORKDIR /proxy

RUN go get github.com/constabulary/gb/...

RUN gb build

EXPOSE 8888/tcp 5683/udp

COPY ./docker/start.sh /start.sh

ENTRYPOINT ["/start.sh"]
