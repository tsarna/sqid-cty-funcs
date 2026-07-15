# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-07-15

### Added

- **`Externs()` — the real signatures, for functy hosts.** `externs.cty` (embedded;
  exposed as opaque bytes via `Externs()` and `ExternsFilename`) declares what both
  functions actually accept, as [functy](https://github.com/tsarna/functy)
  `//functy:extern` declarations. It is never compiled and declares nothing callable; it
  exists so that `help()`, generated documentation, and editor tooling can show what the
  cty metadata cannot:

  - `sqid`'s first argument is a **union** — a single number, or a list of numbers — and
    cty has no union type, so its metadata can only say `dynamic`. It is declared as one
    form per arm.
  - Both functions take an **optional options object** (`alphabet`, `min_length`,
    `blocklist`), which cty can only fake with a variadic — leaving it shapeless. The
    declaration spells the object out attribute by attribute.

  This package does not import functy; the bytes are opaque to it.

  ```go
  parser.RegisterExterns(sqidcty.Externs(), sqidcty.ExternsFilename)
  ```

  The names stay flat: `sqid()`/`unsqid()` are an encode/decode codec pair (like
  `base64encode`/`base64decode`), `sqid()` reads as "make a sqid", and `unsqid()` is its
  plain inverse.

### Changed

- Both functions and every parameter now carry a cty `Description`. The metadata is the
  only documentation a non-functy cty host can see, and the only thing functy's own
  `doc()` reads.

## [0.1.0] - earlier

- Initial release: `sqid` (encode) and `unsqid` (decode), with `alphabet` / `min_length`
  / `blocklist` options.
