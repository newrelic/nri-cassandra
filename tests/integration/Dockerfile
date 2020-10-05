FROM golang:1.9.7 as builder

ARG CGO_ENABLED=0
WORKDIR /go/src/github.com/newrelic/nri-cassandra
COPY . .
RUN make clean compile
RUN wget https://download.newrelic.com/infrastructure_agent/binaries/linux/noarch/nrjmx_linux_1.5.2_noarch.tar.gz \
    && mkdir nrjmx \
    && tar -xvf nrjmx_linux_1.5.2_noarch.tar.gz -C nrjmx

    
FROM maven:3.6-jdk-11
COPY --from=builder /go/src/github.com/newrelic/nri-cassandra/bin /
COPY --from=builder /go/src/github.com/newrelic/nri-cassandra/nrjmx/usr/bin /usr/bin/

CMD ["sleep", "1h"]