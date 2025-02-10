# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.6.2] - 2025-02-10

### Added
- Support for darwin_arm64 build for assets.

## [1.6.0] - 2024-05-30

### Changed
- Update the slack dependent sdk from "github.com/bluele/slack" to "github.com/slack-go/slack" 

## [1.5.0] - 2021-03-022

### Changed
- Fixed README
- Update to use plugin sdk v0.12: adds support for hostname function for use in template

## [1.4.0] - 2021-02-08

### Changed
- Update to use plugin sdk 0.11.0
- Mark --webhook-url argument as secret, so if set via SLACK_WEBHOOK_URL, value will not leak into usage message.
- Q1 '21 handler maintenance:
  - Updated modules (go get -u && go mod tidy)
  - Updated GitHub Actions: Added Lint action
  - Updated build to Go 1.14
  - Fix linter error by having template error go to stdout
  - Change return from sendMessage to capture and embellish any error
  - Output log message that message was sent to Slack
  - Fix TestSendMessage to include template
  - README updates

## [1.3.2] - 2020-05-18

### Changed
- Support for newline (\n) expansion in description template
- README updates

## [1.3.1] - 2020-03-30

### Changed
- Fixed README and CHANGELOG

## [1.3.0] - 2020-03-30

### Added
- Custom output templates using the templating package

## [1.2.0] - 2020-02-12

### Added
- Config options can now be overridden via annotations.

### Changed
- Migrated from go dep to go modules.
- Migrated the build pipeline from Travis to Github Actions.
- Updated license.
- Upgraded to sensu-plugin-sdk.
- Updated the default icon URL.

### Deprecated
- Environment variables prefixed with `SENSU_` are now deprecated and will be
removed in a future release.

## [1.0.3] - 2019-01-09

### Added
- Use SLACK_WEBHOOK_URL envvar for default value of slack_webhook_url.  Use of envvar by default prevents leaking of sensitive credential into system process table via command argument. This is a backwards compatible change, and the --webhook-url argument can still be used as an override for testing purposes.

### Added
- Adds .bonsai.yml.

## [1.0.2] - 2018-12-04

### Added
- Travis post-deploy script generates a sha512 for packages to be sensu asset compatible. 

## [1.0.1] - 2018-11-30

### Changed
- Corrected binary name in help output

## [1.0.0] - 2018-11-30

### Breaking Changes
- Updated sensu-go version to GA RC SHA.

## [0.1.2] - 2018-11-27

### Changed
- Updated the goreleaser file so that the handler is packaged as a sensu compatible asset.

## [0.1.1] - 2018-11-21

### Breaking Changes
- Updated sensu-go version to beta-8 and fixed some breaking changes that
were introduced (Entity.ID -> Entity.Name).

### Removed
- Removed the vendor directory. Dependencies are still managed with Gopkg.toml.

## [0.0.2] - 2018-11-04

Testing Asset goreleaser pipeline.

## [0.0.1] - 2018-08-17

### Added
- More readme instructions

### Changed
- Repo name `slack_handler` to `slack-handler`

## [0.0.0] - 2018-08-17

### Added
- Slack handler
- /vendor
- goreleaser.yml
- travis.yml
- Gopkg.toml
- Gopkg.lock
- LICENSE
- README.md
