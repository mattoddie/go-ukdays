# Contributing

Thanks for your interest in improving go-ukdays. Bug reports, fixes
and documentation improvements are all welcome.

By taking part you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
To report a security problem, follow the [security policy](SECURITY.md)
instead of opening an issue.

## Reporting bugs and requesting features

Search the [existing issues](https://github.com/mattoddie/go-ukdays/issues)
first. If there is nothing similar, open a new issue using one of the
templates. For bugs, the division and dates involved, and what you
expected, help the most. If GOV.UK's list disagrees with the package, link to
the GOV.UK page.

For larger changes, such as new public API, please open an issue to discuss
the idea before writing code.

## Development

You need Go 1.22 or later. The module has no dependencies outside the
standard library.

```sh
git clone https://github.com/mattoddie/go-ukdays.git
cd go-ukdays
go test ./...
```

Before opening a pull request, run:

```sh
go vet ./...
go test -race ./...
golangci-lint run ./...
golangci-lint fmt --diff ./...
```

[golangci-lint](https://golangci-lint.run/welcome/install/) v2 is used for
linting and formatting; its configuration is in `.golangci.yml`.

Changes to the calendar or date handling should be checked with the fuzz
tests as well:

```sh
go test -run '^$' -fuzz FuzzWorkingDays -fuzztime 1m .
go test -run '^$' -fuzz FuzzParseDate -fuzztime 30s .
```

If fuzzing finds a failure, Go saves the input under `testdata/fuzz/`. Commit
that file with your fix so it becomes a regression test.

## Bank holiday data

`data/bank-holidays.json` is GOV.UK's list, unchanged. Don't edit it by hand;
refresh it with:

```sh
go generate ./...
```

A scheduled workflow does this every Monday and opens a pull request when the
list changes. Pull requests opened by GitHub Actions don't start the other
checks, so close and reopen it before merging.

The rules in `rules.go` must keep matching every published year;
`TestRulesMatchPublished` checks this. When GOV.UK announces a one-off
holiday, add its date to `publishedOnly` in `rules_test.go`.

## Pull requests

- Keep each pull request focused on one change.
- Add or update tests for any change in behaviour.
- Document new exported identifiers with doc comments, and add an example
  test when it helps.
- Add an entry under "Unreleased" in [CHANGELOG.md](CHANGELOG.md) for
  user-visible changes.
- Use [Conventional Commits](#commit-messages) for commit messages and the
  pull request title.

CI runs the tests on Linux, macOS and Windows, along with linting, fuzzing,
govulncheck and CodeQL. All checks must pass before merging.

### Commit messages

Commit messages and pull request titles follow
[Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).
Pull requests are squash-merged, so the title becomes the commit on `main`;
CI checks that it matches.

```text
<type>(<optional scope>)<optional !>: <description>
```

For example `fix(rules): match GOV.UK for 2030`,
`feat(cmd): add a --year flag to ukdays`,
`chore(data): update bank holidays from GOV.UK` or
`ci(deps): bump the github-actions group with 2 updates`.

- **Types:** `feat`, `fix`, `docs`, `refactor`, `perf`, `test`, `build`
  (dependencies and toolchain), `ci` (workflows), `chore` (anything else,
  such as `chore(release): v1.2.0`), `revert` and `style`.
- **Description:** imperative mood, lower-case first word, no trailing full
  stop. Keep the whole subject to 72 characters or fewer.
- **Breaking changes:** add `!` after the type or scope, for example
  `feat(calendar)!: …`, and a `BREAKING CHANGE: …` footer in the body (or the
  pull request description). List the change under "Breaking changes" in the
  changelog too.
- **Body:** explain why when it isn't obvious, and reference issues with
  `Fixes #123`.

## Compatibility

This module follows [semantic versioning](https://semver.org/) and Go's
[module compatibility rules](https://go.dev/blog/module-compatibility).
Until v1.0.0 the API may still change between minor versions; breaking changes
are listed in the changelog. Changes to the JSON or text formats should be
discussed in an issue first.

## Licence

By contributing, you agree that your contributions will be licensed under the
[MIT License](LICENSE).
