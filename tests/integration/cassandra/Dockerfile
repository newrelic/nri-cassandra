FROM cassandra:3.11

ENV JVM_OPTS="-Xmn100m -Xmx384m -Xms128m"
ENV LOCAL_JMX="no"
RUN sed -i 's/# JVM_OPTS=\"$JVM_OPTS -Djava.rmi.server.hostname=<public name>\"/JVM_OPTS=\"$JVM_OPTS -Djava.rmi.server.hostname=cassandra\"/g' /etc/cassandra/cassandra-env.sh \
    && sed -i 's/JVM_OPTS=\"$JVM_OPTS -Dcom.sun.management.jmxremote.authenticate=true\"/JVM_OPTS=\"$JVM_OPTS -Dcom.sun.management.jmxremote.authenticate=false\"/g' /etc/cassandra/cassandra-env.sh \
