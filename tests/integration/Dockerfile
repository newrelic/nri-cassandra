FROM golang:1.11 as builder
ARG CGO_ENABLED=0
WORKDIR /go/src/github.com/newrelic/nri-cassandra
COPY . .
RUN make clean compile

FROM gradle:6.6.1-jdk11 as jmxbuilder
RUN git clone https://github.com/newrelic/nrjmx.git \
    && cd nrjmx \
    && gradle package

FROM gradle:6.6.1-jdk11
COPY --from=builder /go/src/github.com/newrelic/nri-cassandra/bin /
COPY --from=jmxbuilder /home/gradle/nrjmx/bin /usr/bin/
COPY --from=jmxbuilder /home/gradle/nrjmx/build/distributions/nrjmx-*-noarch.jar /usr/bin/nrjmx.jar

CMD ["sleep", "1h"]
