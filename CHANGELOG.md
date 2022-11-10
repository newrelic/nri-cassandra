# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## 2.11.0 (2022-11-10)
### Changed
- Updated gojmx library to [v2.3.0](https://github.com/newrelic/nrjmx/releases/tag/v2.3.0)

## 2.11.0 (2022-09-06)
### Changed
- Updated gojmx library to [v2.2.2](https://github.com/newrelic/nrjmx/releases/tag/v2.2.2)
- Optimisation of number of JMX queries by removing wildcards when possible and fetching only the required attributes.
### Added
- `ENABLE_INTERNAL_STATS` configuration option. When this option is enabled and the integration is running in verbose mode it will output in the logs nrjmx internal query statistics. This will be handy when troubleshooting performance issues.
- [BETA] Added long-running mode (LONG_RUNNING config option). When running in this mode the RMI connection will be reused instead of creating a new one every collection.
- [BETA] Added MBean filtering configuration.
```yaml
METRICS_FILTER: >-
  exclude:
    - "*"
  include:
    - client.connectedNativeClients
    - db.droppedRangeSliceMessagesPerSecond
    - db.tombstoneScannedHistogram999thPercentile
```

### Fixed
- Issue when JMX connection is opened unnecessarily in Inventory collection mode.

## 2.10.2 (2022-06-10)
### Changed
- Use Go 1.18 to compile the integration
- Bump dependencies: 
  `github.com/newrelic/infra-integrations-sdk` to version `3.7.3+incompatible`
  `github.com/stretchr/testify` to version `1.7.2`
### Added
* Cassandra Logging Template File (#87)

## 2.10.1 (2022-05-24)
### Changed
* Updated gojmx library to v2.0.2

## 2.10.0 (2022-04-11)
### Added
* migrate to gojmx by @cristianciutea in https://github.com/newrelic/nri-cassandra/pull/82

## 2.9.1 (2021-10-20)
### Added
Added support for more distributions:
- Debian 11
- Ubuntu 20.10
- Ubuntu 21.04
- SUSE 12.15
- SUSE 15.1
- SUSE 15.2
- SUSE 15.3
- Oracle Linux 7
- Oracle Linux 8

## 2.9.0 (2021-08-30)
### Changed
- Moved default config.sample to [V4](https://docs.newrelic.com/docs/create-integrations/infrastructure-integrations-sdk/specifications/host-integrations-newer-configuration-format/), added a dependency for infra-agent version 1.20.0

Please notice that old [V3](https://docs.newrelic.com/docs/create-integrations/infrastructure-integrations-sdk/specifications/host-integrations-standard-configuration-format/) configuration format is deprecated, but still supported.

## 2.8.1 (2021-06-10)
### Changed
- ARM support added

## 2.8.0 (2021-05-04)
### Changed
- Update Go to v1.16.
- Migrate to Go Modules
- Update Infrastracture SDK to v3.6.7.
- Update other dependecies.

## 2.7.0 (2021-03-30)
### Changed
- Upgraded to Infrastructure SDK v3.6.6 which has a fix for handling warning messages from [nrjmx](https://github.com/newrelic/nrjmx).

## 2.6.0 (2020-11-19)
### Changed
- Enable JMX connections with SSL/TLS.
- Updated the configuration sample to include the hostname for inventory required
  for entity naming.

## 2.5.0 (2020-09-07)
### Changed
- Add hinted handoff manager metrics

## 2.4.2 (2020-04-16)
### Changed
- Upgraded to SDK v3.6.3. This version provides improvements for jmx error handling.
- Handle jmx client error. Stop query execution when jmx tool has stopped. 

## 2.4.1 (2020-04-07)
### Changed
- Upgraded to SDK v3.6.2. This version provides a fix for an issue which caused some debug logs to be lost.

## 2.4.0 (2019-11-22)
### Changed
- Renamed the integration executable from nr-cassandra to nri-cassandra in order to be consistent with the package naming. **Important Note:** if you have any security module rules (eg. SELinux), alerts or automation that depends on the name of this binary, these will have to be updated.

## 2.3.0 (2019-11-18)
### Added
- Add nrjmx version dependency to 1.5.2, so jmxterm can be bundled within
  package.
- Upgraded to SDK v3.5.0. This version provides improvements for jmx support and also better error handling.

## 2.2.0 (2019-04-29)
### Added
- Upgraded to SDK v3.1.5. This version implements [the aget/integrations
  protocol v3](https://github.com/newrelic/infra-integrations-sdk/blob/cb45adacda1cd5ff01544a9d2dad3b0fedf13bf1/docs/protocol-v3.md),
  which enables [name local address replacement](https://github.com/newrelic/infra-integrations-sdk/blob/cb45adacda1cd5ff01544a9d2dad3b0fedf13bf1/docs/protocol-v3.md#name-local-address-replacement).
  and could change your entity names and alarms. For more information, refer
  to:
  
  - https://docs.newrelic.com/docs/integrations/integrations-sdk/file-specifications/integration-executable-file-specifications#h2-loopback-address-replacement-on-entity-names
  - https://docs.newrelic.com/docs/remote-monitoring-host-integration://docs.newrelic.com/docs/remote-monitoring-host-integrations 

## 2.1.0 (2019-04-08)
### Added
- Remote monitoring option. It enables monitoring multiple instances, 
  more information can be found at the [official documentation page](https://docs.newrelic.com/docs/remote-monitoring-host-integrations).

## 2.0.3 (2018-12-04)
### Added
- Fix bug that made db.keyspace, db.columnFamily,db.keyspaceAndColumnFamily to be filled with default values.

## 2.0.2 (2018-09-12)
### Added
- Fix `integrationVersion` field for cassandra integration.

## 2.0.0 (2018-09-6)
### Added
- Updated SDK to v3.
- Cache collision bug fixed.

## 1.2.0 (2018-07-11)
### Added
- Add threadpool and histogram metrics

## 1.1.0 (2017-10-16)
### Added
- Set column families limit as a configurable value
- Set query timeout as a configurable value

## 0.1.0
### Added
- Initial release, which contains inventory and metrics data
