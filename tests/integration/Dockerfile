FROM golang:1.22-bookworm as builder

ARG CGO_ENABLED=0

WORKDIR /go/src/github.com/newrelic/nri-cassandra
COPY . .
RUN make clean compile
WORKDIR /

# Install latest version of nrjmx
# Package installation requires a version of nri-jmx grater than a value so this is why we test with the latest one.
RUN curl -fsSL https://download.newrelic.com/infrastructure_agent/gpg/newrelic-infra.gpg | gpg --dearmor -o /etc/apt/trusted.gpg.d/newrelic-infra.gpg && \
  echo "deb https://download.newrelic.com/infrastructure_agent/linux/apt bookworm main" | tee -a /etc/apt/sources.list.d/newrelic-infra.list && \
  apt-get update && \
  apt-get install nrjmx

FROM maven:3.6-jdk-11
COPY --from=builder /go/src/github.com/newrelic/nri-cassandra/bin /
COPY --from=builder /usr/bin/nrjmx /usr/bin/nrjmx
COPY --from=builder /usr/lib/nrjmx/nrjmx.jar /usr/lib/nrjmx/nrjmx.jar

COPY ./tests/integration/cassandra/certs /certs

CMD ["sleep", "1h"]
