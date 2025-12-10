# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

Unreleased section should follow [Release Toolkit](https://github.com/newrelic/release-toolkit#render-markdown-and-update-markdown)

## Unreleased

## v2.16.1 - 2025-12-10

### ⛓️ Dependencies
- Updated golang patch version to v1.25.5

## v2.16.0 - 2025-11-05

### 🛡️ Security notices
- Updated golang version to v1.25.3

### 🚀 Enhancements
- Updated nrjmx-fips patch version to v2.10.2

## v2.15.0 - 2025-09-03

### 🚀 Enhancements
- Add FIPS compliant packages

## v2.14.7 - 2025-08-13

### ⛓️ Dependencies
- Updated golang patch version to v1.24.6

## v2.14.6 - 2025-07-24

### 🐞 Bug fixes
- Rereleasing due to bad release

## v2.14.5 - 2025-07-10

### ⛓️ Dependencies
- Updated golang patch version to v1.24.5

## v2.14.4 - 2025-07-02

### ⛓️ Dependencies
- Updated golang version to v1.24.4

## v2.14.3 - 2025-02-05

### ⛓️ Dependencies
- Updated golang patch version to v1.23.6

## v2.14.2 - 2025-01-29

### ⛓️ Dependencies
- Updated golang patch version to v1.23.5

## v2.14.1 - 2024-12-04

### ⛓️ Dependencies
- Updated golang patch version to v1.23.4

## v2.14.0 - 2024-10-09

### dependency
- Upgrade go to 1.23.2

### 🚀 Enhancements
- Upgrade integrations SDK so the interval is variable and allows intervals up to 5 minutes

## v2.13.10 - 2024-09-11

### ⛓️ Dependencies
- Updated golang version

## v2.13.9 - 2024-08-14

### ⛓️ Dependencies
- Updated golang version to v1.22.6

## v2.13.8 - 2024-07-03

### ⛓️ Dependencies
- Updated golang version to v1.22.5

## v2.13.7 - 2024-06-26

### ⛓️ Dependencies
- Updated golang version to v1.22.4
- Updated github.com/newrelic/nrjmx/gojmx digest

## v2.13.6 - 2024-05-08

### ⛓️ Dependencies
- Updated golang version to v1.22.3

## v2.13.5 - 2024-04-10

### ⛓️ Dependencies
- Updated golang version

## v2.13.4 - 2024-02-22

### 🐞 Bug fixes
- Update publish schema

## v2.13.3 - 2024-02-21

### 🐞 Bug fixes
- Fix MBeanAttribute names to match the actual attribute names for metrics: 'db.threadpool.nativeTransportRequestActiveTasks', 'db.threadpool.nativeTransportRequestCompletedTasks', and 'db.threadpool.nativeTransportRequestPendingTasks'

### ⛓️ Dependencies
- Updated github.com/newrelic/infra-integrations-sdk to v3.8.2+incompatible

## v2.13.2 - 2023-08-02

### ⛓️ Dependencies
- Updated golang to v1.20.7

## v2.13.1 - 2023-07-19

### ⛓️ Dependencies
- Updated golang version to 1.20

## 2.13.0 (2023-05-11)
### Changed
- Disable CGO
### Added
- Include NTR threadpool metrics

## 2.12.0 (2023-03-08)
### Changed
- Updated gojmx library to [v2.3.2](https://github.com/newrelic/nrjmx/releases/tag/v2.3.2)
- Upgraded Go version to 1.19
- General upgrade of all dependencies

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
