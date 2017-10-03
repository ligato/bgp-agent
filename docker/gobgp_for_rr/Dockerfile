FROM golang

MAINTAINER Claudio D. Gasparini <claudio.gasparini@pantheon.sk>

EXPOSE 179

RUN go get github.com/osrg/gobgp/gobgp
RUN go get github.com/osrg/gobgp/gobgpd

RUN mkdir /etc/gpbgp

WORKDIR /go

CMD gobgpd -f /etc/gobgp/gobgp.conf -p > /etc/gobgp/log