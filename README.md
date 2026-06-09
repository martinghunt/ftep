# ftep

Finding things in the ENA portal.

Currently supported: run, experiment, sample, study/project, assembly accessions.

This repository was developed with substantial coding assistance from
[OpenAI Codex](https://openai.com/codex), which helped with implementation,
refactoring, tests, documentation, and benchmarking under human direction and review.


## Install

The simplest way to install `ftep` is to download the latest prebuilt binary from the GitHub releases page:

- https://github.com/martinghunt/ftep/releases/latest

Choose the archive or binary matching your OS and CPU architecture.

After installing, check the version with:

```
ftep --version
```

If you want to build locally instead:

```
./build.sh
```

That builds `ftep` for the current OS and architecture into `./build/ftep` or `./build/ftep.exe`.
Local builds report version `dev` unless you pass an explicit release version.

For a cross-platform release build:

```
./build.sh --release --version v1.2.3
```


## Synopsis

Get metadata for sample `SAMN05276490` in (default) TSV format:
```
ftep search -a SAMN05276490
```

Get metadata for accessions (one per line, must all be same type eg runs, samples etc)
in the file `acc.txt`:
```
ftep search -f acc.txt
```

Get metadata for sample `SAMN05276490` in JSON format:
```
ftep search -a SAMN05276490 --outfmt json
```

Get metadata for sample `SAMN05276490` as an aligned table:
```
ftep search -a SAMN05276490 --outfmt table
```

Get all available metadata for sample `SAMN05276490`:
```
ftep search -a SAMN05276490 -c ALL
```

Get runs for sample `SAMN05276490`:
```
ftep search -a SAMN05276490 --level run
```

Get metadata for study/project `PRJEB1787`:
```
ftep search -a PRJEB1787
```

Get samples for study/project `PRJEB1787`:
```
ftep search -a PRJEB1787 --level sample
```

Get runs for study/project `PRJEB1787`:
```
ftep search -a PRJEB1787 --level run
```

Get a FASTQ download manifest for sample `SAMN05276490`:
```
ftep reads -a SAMN05276490
```

Get the FASTQ download manifest as an aligned table:
```
ftep reads -a SAMN05276490 --outfmt table
```

Print `wget` commands to download FASTQs for sample `SAMN05276490`:
```
ftep reads -a SAMN05276490 --outfmt wget
```

Print MD5 checksum lines for those FASTQs:
```
ftep reads -a SAMN05276490 --outfmt md5
```

Open sample `SAMN05276490` in the ENA browser:
```
ftep open SAMN05276490
```

Print the ENA browser URL for run `SRR3675520`:
```
ftep open SRR3675520 --print-url
```

List available ENA data types and whether `ftep search` supports them, with
supported types first:
```
ftep get_fields --outfmt table
```

List available fields for ENA data type `read_run`:
```
ftep get_fields read_run
```

Get metadata for study accession `ERP001736`:
```
ftep search -a ERP001736
```

Get metadata for run `SRR3675520`:
```
ftep search -a SRR3675520
```

Get metadata for assembly `GCA_000195955.2`:
```
ftep search -a GCA_000195955.2
```


## Go library

Import the module and use the ENA client directly:

```go
package main

import (
	"context"
	"fmt"

	"github.com/martinghunt/ftep"
)

func main() {
	client := ftep.NewClient()
	results, err := client.Search(context.Background(), ftep.SearchOptions{
		Accessions: []string{"SAMN05276490"},
		Fields:     []string{"DEFAULT"},
		Level:      ftep.AccessionTypeRun,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(results[0].Records)
}
```


## For developers

Releases are made from Git tags. The GitHub Actions release workflow runs when a tag matching `v*.*.*` is pushed. It runs the tests, builds binaries for Darwin, Linux, and Windows on amd64 and arm64, then uploads the archives to the GitHub release.

Before tagging, run:

```
go test ./...
./build.sh
```

Then create and push the release tag:

```
git tag -a v1.2.3 -m "ftep v1.2.3"
git push origin main
git push origin v1.2.3
```

For a local check of the full release matrix:

```
./build.sh --release --version v1.2.3
```
