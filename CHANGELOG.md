# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Ability to manage accounts

## [1.3.1] - 2023-04-30

### Fixed

- Message delay in case there are other accounts that may be used to sent it

## [1.3.0] - 2023-04-29

### Added

- Ability to manage own categories

## [1.2.1] - 2023-04-29

### Fixed

- Sent message id parsing from response

## [1.2.0] - 2023-04-29

### Added

- Ability to start bot without sender

### Changed

- `SLEEP_DURATION` environment variable is renamed to `SENDER_SLEEP_DURATION`

## [1.1.0] - 2023-04-23

### Added

- Ability to delete a message

## [1.0.0] - 2023-04-20

### Added

- Notify about sent message

## [0.20.1] - 2023-04-18

### Fixed

- Missed `sender` start invocation

## [0.20.0] - 2023-04-17

### Changed

- DI to `uber-go/fx`

## [0.19.0] - 2023-04-16

### Added

- Last access time to user state

## [0.18.5] - 2023-04-16

### Fixed

- Multiple files in attachment

## [0.18.4] - 2023-04-15

### Fixed

- Category selection

## [0.18.3] - 2023-04-14

### Fixed

- Setting message priority
- Choosing parent cateogory when creating a message

## [0.18.2] - 2023-04-14

### Fixed

- Saving reauthorized account token in the state

## [0.18.1] - 2023-04-13

### Removed

- Unused deprecated user state fields

### Fixed

- String array field reading from state form

## [0.18.0] - 2023-04-13

### Added

- Multiple portal accounts support

## [0.17.0] - 2023-04-08

### Added

- High message priority to send a message as soon as possible

### Changed

- Login required to send a message

## [0.16.0] - 2023-04-03

### Changed

- Keep categories order equal to the configuration
- Show two category buttons per row

## [0.15.0] - 2023-03-28

### Added

- Sort messages by creation date

## [0.14.0] - 2023-03-27

### Added

- Retry requests to portal

## [0.13.3] - 2023-03-27

### Fixed

- Message fail description to contain root cause of the error

## [0.13.2] - 2023-03-26

### Fixed

- Next try time calculation

## [0.13.1] - 2023-03-26

### Fixed

- Debug logging for polling and adding a message to queue

## [0.13.0] - 2023-03-25

### Added

- User rate limited status in `/status` command output

## [0.12.5] - 2023-03-25

### Fixed

- Reduce log size when error is returned

## [0.12.4] - 2023-03-25

### Fixed

- Add logging when reading user state

## [0.12.3] - 2023-03-25

### Fixed

- Read a single document when polling a message
- Add logging when saving user state

## [0.12.2] - 2023-03-23

### Fixed

- Sort files by width and height instead of file size

## [0.12.1] - 2023-03-22

### Fixed

- Do not increase try count for rate limited requests

## [0.12.0] - 2023-03-21

### Added

- `/file_id` command to return photos file identifiers

## [0.11.0] - 2023-03-21

### Changed

- Store `fileId` instead of `fileUrl` to download files later

## [0.10.0] - 2023-03-20

### Added

- Version info to `/start` command output

### Changed

- Reply about database rate limits on `/start` and `/status` commands

## [0.9.2] - 2023-03-20

### Fixed

- Endless loop when can't download files or another error

## [0.9.1] - 2023-03-19

### Fixed

- Error when sender fails to download files

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