# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-10-12

### Added

- `Calendar` for England and Wales, Scotland and Northern Ireland, with
  `IsWorkingDay`, `NextWorkingDay`, `PreviousWorkingDay`, `AddWorkingDays`,
  `WorkingDaysBetween`, `Holidays`, `HolidaysBetween` and `HolidayOn`.
- Built-in copy of GOV.UK's bank holidays for 2019 to 2028, with `Fetch` and
  `Load` for the latest list.
- Holidays from 2007 onwards for years outside GOV.UK's list, worked out from
  the standing rules and tested against every published year.
- `Date`, a calendar date type with JSON support.
- The `ukdays` command-line tool.

[Unreleased]: https://github.com/mattoddie/go-ukdays/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mattoddie/go-ukdays/releases/tag/v0.1.0
