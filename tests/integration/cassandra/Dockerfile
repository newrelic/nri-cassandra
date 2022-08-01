FROM cassandra:3.11

ENV JVM_OPTS="-Xmn100m -Xmx384m -Xms128m"
ENV LOCAL_JMX="no"

COPY ./tests/integration/cassandra/certs /opt/cassandra/conf/certs
COPY ./tests/integration/cassandra/auth/jmxremote.password /etc/cassandra/jmxremote.password
COPY ./tests/integration/cassandra/auth/jmxremote.access /opt/java/openjdk/lib/management/jmxremote.access

COPY ./tests/integration/cassandra/cassandra.yaml /etc/cassandra/cassandra.yaml

RUN sed -i 's/# JVM_OPTS=\"$JVM_OPTS -Djava.rmi.server.hostname=<public name>\"//g' /etc/cassandra/cassandra-env.sh \
    && sed -i 's/JVM_OPTS=\"$JVM_OPTS -Dcom.sun.management.jmxremote.authenticate=true\"//g' /etc/cassandra/cassandra-env.sh

RUN echo 'JVM_OPTS="$JVM_OPTS $EXTRA_JVM_OPTS"' >> /etc/cassandra/cassandra-env.sh