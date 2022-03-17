FROM golang:1.16 as builder

ARG CGO_ENABLED=0
ARG NRJMX_VERSION=2.0.1

WORKDIR /go/src/github.com/newrelic/nri-cassandra
COPY . .
RUN make clean compile
WORKDIR /

ADD https://github.com/newrelic/nrjmx/releases/download/v${NRJMX_VERSION}/nrjmx_linux_${NRJMX_VERSION}_noarch.tar.gz ./

RUN tar -xzf nrjmx_linux_${NRJMX_VERSION}_noarch.tar.gz

FROM maven:3.6-jdk-11
COPY --from=builder /go/src/github.com/newrelic/nri-cassandra/bin /
COPY --from=builder /usr/bin/nrjmx /usr/bin/nrjmx
COPY --from=builder /usr/lib/nrjmx/nrjmx.jar /usr/lib/nrjmx/nrjmx.jar

CMD ["sleep", "1h"]
