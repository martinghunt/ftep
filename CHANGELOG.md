# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-06-09

### Added
- Add `ichsm identify` to classify accessions, show normalized forms, and report ENA/NCBI search support.
- Add a weekly GitHub Actions live smoke test for the public ENA and NCBI endpoints.
- Add ENA `sequence`, `coding`, `tsa_set`, and `tls_set` search support.
- Add support for WGS/TSA/TLS short set IDs and component-shaped sequence accessions.
- Add `ichsm search --source auto|ena|ncbi`, defaulting to ENA first with NCBI fallback.
- Add `ichsm open --source auto|ena|ncbi` with NCBI browser URLs for NCBI-only or forced NCBI accessions.
- Add NCBI E-utilities metadata fallback for `GCF_`, RefSeq nucleotide, and RefSeq protein accessions.
- Add `ichsm search --api-key` and `--email`, defaulting to `NCBI_API_KEY` and `NCBI_EMAIL`, for NCBI requests.
- Add support for WGS master accessions such as `AGQU00000000.1`.
- Add support for ENA study/project accessions, including `PRJEB`, `PRJDB`, `PRJNA`, `ERP`, `DRP`, and `SRP` accessions.
- Add `ichsm search --level` to choose study, sample, run, or assembly output level where supported by the input accession type.
- Add `ichsm reads` to print FASTQ download manifests, URLs, `wget` commands, `curl` commands, or MD5 checksum lines.
- Add `ichsm open` to open an accession in the ENA browser or print its browser URL.
- Let `ichsm get_fields` list available ENA data types and whether `ichsm search` supports them when no data type is supplied.
- Add aligned table output for `ichsm search`, `ichsm reads`, and `ichsm get_fields`.

### Changed
- Refresh CLI help and Go documentation to describe ENA and NCBI support.
- Rename the project, CLI, Go module path, and release artifacts from `ftep` to `ichsm`.
- Use `ichsm reads --outfmt` for output selection, matching `ichsm search` and `ichsm get_fields`.

### Removed
- Remove `ichsm search --s2r`; use `ichsm search --level run` instead.

## [0.1.0] - 2026-05-29

Release `v0.1.0`, before changelog tracking started in this file.

[Unreleased]: https://github.com/martinghunt/ichsm/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/martinghunt/ichsm/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/martinghunt/ichsm/releases/tag/v0.1.0
