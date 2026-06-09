# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Add ENA `sequence`, `coding`, `tsa_set`, and `tls_set` search support.
- Add support for WGS/TSA/TLS short set IDs and component-shaped sequence accessions.
- Add `ftep search --source auto|ena|ncbi`, defaulting to ENA first with NCBI fallback.
- Add `ftep open --source auto|ena|ncbi` with NCBI browser URLs for NCBI-only or forced NCBI accessions.
- Add NCBI E-utilities metadata fallback for `GCF_`, RefSeq nucleotide, and RefSeq protein accessions.
- Add `ftep search --api-key` and `--email`, defaulting to `NCBI_API_KEY` and `NCBI_EMAIL`, for NCBI requests.
- Add support for WGS master accessions such as `AGQU00000000.1`.
- Add support for ENA study/project accessions, including `PRJEB`, `PRJDB`, `PRJNA`, `ERP`, `DRP`, and `SRP` accessions.
- Add `ftep search --level` to choose study, sample, run, or assembly output level where supported by the input accession type.
- Add `ftep reads` to print FASTQ download manifests, URLs, `wget` commands, `curl` commands, or MD5 checksum lines.
- Add `ftep open` to open an accession in the ENA browser or print its browser URL.
- Let `ftep get_fields` list available ENA data types and whether `ftep search` supports them when no data type is supplied.
- Add aligned table output for `ftep search`, `ftep reads`, and `ftep get_fields`.

### Changed
- Use `ftep reads --outfmt` for output selection, matching `ftep search` and `ftep get_fields`.

### Removed
- Remove `ftep search --s2r`; use `ftep search --level run` instead.

## [0.1.0] - 2026-05-29

Release `v0.1.0`, before changelog tracking started in this file.

[Unreleased]: https://github.com/martinghunt/ftep/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/martinghunt/ftep/releases/tag/v0.1.0
