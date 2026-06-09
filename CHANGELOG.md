# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Add support for ENA study/project accessions, including `PRJEB`, `PRJDB`, `PRJNA`, `ERP`, `DRP`, and `SRP` accessions.
- Add `ftep search --level` to choose study, sample, run, or assembly output level where supported by the input accession type.
- Add `ftep reads` to print FASTQ download manifests, URLs, `wget` commands, `curl` commands, or MD5 checksum lines.
- Add `ftep open` to open an accession in the ENA browser or print its browser URL.

### Removed
- Remove `ftep search --s2r`; use `ftep search --level run` instead.

## [0.1.0] - 2026-05-29

Release `v0.1.0`, before changelog tracking started in this file.

[Unreleased]: https://github.com/martinghunt/ftep/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/martinghunt/ftep/releases/tag/v0.1.0
