FROM golang:1.20 as builder-cassandra
COPY . /go/src/github.com/newrelic/nri-cassandra/
RUN cd /go/src/github.com/newrelic/nri-cassandra && \
    make && \
    strip ./bin/nri-cassandra

FROM maven:3-jdk-11 as builder-jmx
RUN git clone https://github.com/newrelic/nrjmx && \
    cd nrjmx && \
    mvn package -DskipTests -P \!deb,\!rpm,\!test,\!tarball

FROM newrelic/infrastructure:latest
ENV NRIA_IS_FORWARD_ONLY true
ENV NRIA_K8S_INTEGRATION true

COPY --from=builder-cassandra /go/src/github.com/newrelic/nri-cassandra/bin/nri-cassandra /nri-sidecar/newrelic-infra/newrelic-integrations/bin/nri-cassandra
COPY --from=builder-jmx /nrjmx/bin /usr/bin/

RUN apk update && apk add openjdk8-jre
USER 1000

