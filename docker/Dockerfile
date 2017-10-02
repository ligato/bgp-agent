FROM golang

MAINTAINER Claudio D. Gasparini <claudio.gasparini@pantheon.sk>

EXPOSE 179

RUN apt-get update

RUN go get github.com/osrg/gobgp/gobgp
RUN go get github.com/osrg/gobgp/gobgpd

RUN mkdir /etc/gpbgp

COPY gobgp-client-in-docker.conf /etc/gobgp/gobgp-client-in-docker.conf
COPY gobgp-client-in-host.conf /etc/gobgp/gobgp-client-in-host.conf
COPY gobgp-benchmark.conf /etc/gobgp/gobgp-benchmark.conf

ENV GOBGP_CONFIG gobgp-client-in-host.conf

CMD gobgpd -f /etc/gobgp/$GOBGP_CONFIG --log-level debug
