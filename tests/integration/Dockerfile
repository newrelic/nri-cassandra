FROM golang:1.11 as builder
ARG CGO_ENABLED=0
WORKDIR /go/src/github.com/newrelic/nri-cassandra
COPY . .
RUN make clean compile

FROM maven:3.6-jdk-7 as jmxbuilder
RUN git clone https://github.com/newrelic/nrjmx.git \
    && cd nrjmx \
    && mvn clean package -DskipTests -P \!deb,\!rpm,\!test,\!tarball

FROM maven:3.6-jdk-7
COPY --from=builder /go/src/github.com/newrelic/nri-cassandra/bin /
COPY --from=jmxbuilder /nrjmx/bin /usr/bin/

CMD ["sleep", "1h"]