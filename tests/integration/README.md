nri-cassandra integration tests:

- Configuring TLS JMX connection.

The password for certificate located in ./tests/integration/cassandra/certs/ is 'cassandra'

In order to configure Cassandra server to enable SSL JMX connection edit cassandra-env.sh:

JVM_OPTS="$JVM_OPTS -Dcom.sun.management.jmxremote.ssl=true "
JVM_OPTS="$JVM_OPTS -Dcom.sun.management.jmxremote.ssl.need.client.auth=true "
JVM_OPTS="$JVM_OPTS -Dcom.sun.management.jmxremote.registry.ssl=true "
JVM_OPTS="$JVM_OPTS -Dcom.sun.management.jmxremote=true "
JVM_OPTS="$JVM_OPTS -Djavax.net.ssl.keyStore=/opt/cassandra/conf/certs/cassandra.keystore  "
JVM_OPTS="$JVM_OPTS -Djavax.net.ssl.keyStorePassword=cassandra "
JVM_OPTS="$JVM_OPTS -Djavax.net.ssl.trustStore=/opt/cassandra/conf/certs/cassandra.truststore "
JVM_OPTS="$JVM_OPTS -Djavax.net.ssl.trustStorePassword=cassandra "

Inside the tests you can use the helper function that configures a docker container:

compose := testutils.ConfigureSSLCassandraDockerCompose()
err := testutils.RunDockerCompose(s.compose)
