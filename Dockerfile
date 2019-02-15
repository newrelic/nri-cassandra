FROM golang:1.10 as builder-cassandra
RUN go get -d github.com/newrelic/nri-cassandra/... && \
    cd /go/src/github.com/newrelic/nri-cassandra && \
    make && \
    strip ./bin/nr-cassandra

FROM golang:1.10 as builder-jmx
RUN go get -d github.com/newrelic/nri-jmx/... && \
    cd /go/src/github.com/newrelic/nri-jmx && \
    make && \
    ls /go/src/github.com/newrelic/nri-jmx/bin

FROM newrelic/infrastructure:latest
COPY --from=builder-cassandra /go/src/github.com/newrelic/nri-cassandra/bin/nr-cassandra /var/db/newrelic-infra/newrelic-integrations/bin/nr-cassandra
COPY --from=builder-cassandra /go/src/github.com/newrelic/nri-cassandra/cassandra-definition.yml /var/db/newrelic-infra/newrelic-integrations/definition.yaml
COPY --from=builder-jmx /go/src/github.com/newrelic/nri-jmx/bin/nr-jmx /usr/bin/nrjmx
