# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

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
