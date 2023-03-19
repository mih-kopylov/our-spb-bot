# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.1] - 2023-03-18

### Fixed

- Fix error when sender fails to download files

## [0.9.0] - 2023-03-18

### Added

- Additional categories

### Changed

- Show if user is authorized on the portal or not
- Store username on `/start` command

## [0.8.0] - 2023-03-18

### Added

- `/reset_status` command to reset failed messages status

### Changed

- Do not try sending a message if user is rate limited

## [0.7.1] - 2023-03-12

### Fixed

- Wait before next try if sender fails to poll a message

## [0.7.0] - 2023-03-12

### Added

- Quick links to `/message` command in other commands output

## [0.6.0] - 2023-03-12

### Changed

- Keep only 3 statuses: `created`, `failed`, `awaiting_authorization`
- Show the statuses info in `/status` command
- Reduce number of readings from database to avoid hitting free rate limit

## [0.5.0] - 2023-03-12

### Added

- Additional categories

## [0.4.2] - 2023-03-11

### Fixed

- Error response message verification

## [0.4.1] - 2023-03-11

### Fixed

- Keep message retryable when too many requests a day is sent

## [0.4.0] - 2023-03-11

### Added

- Fail description field to the message

## [0.3.0] - 2023-03-11

### Added

- Firebase to store user state and messages queue

## [0.2.0] - 2023-03-06

### Added

- `docker-compose.yml` template

### Fixed

- Docker image entrypoint

## [0.1.0] - 2023-03-05

### Added

- Ability to send a message
- Static hardcoded categories
- In-memory queue