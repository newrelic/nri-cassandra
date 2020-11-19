# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

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
