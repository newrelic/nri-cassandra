FROM golang:1.10 as builder-cassandra
RUN go get -d github.com/newrelic/nri-cassandra/... && \
    cd /go/src/github.com/newrelic/nri-cassandra && \
    make && \
    strip ./bin/nr-cassandra

FROM maven:3-jdk-11 as builder-jmx
RUN git clone https://github.com/newrelic/nrjmx && \
    cd nrjmx && \
    mvn clean package -P \!deb,\!rpm 

FROM newrelic/infrastructure:latest
COPY --from=builder-cassandra /go/src/github.com/newrelic/nri-cassandra/bin/nr-cassandra /var/db/newrelic-infra/newrelic-integrations/bin/nr-cassandra
COPY --from=builder-cassandra /go/src/github.com/newrelic/nri-cassandra/cassandra-definition.yml /var/db/newrelic-infra/newrelic-integrations/definition.yaml
COPY --from=builder-jmx nrjmx/bin/nrjmx /usr/bin/nrjmx
COPY --from=builder-jmx nrjmx/target/nrjmx-*-jar-with-dependencies.jar /usr/lib/nrjmx/nrjmx.jar
RUN apk update && apk add openjdk7-jre