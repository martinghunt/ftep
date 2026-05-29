# ftep

Finding things in the ENA portal.

Currently supported: run, sample, assembly accessions.


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

Get all available metadata for sample `SAMN05276490`:
```
ftep search -a SAMN05276490 -c ALL
```

Get runs for sample `SAMN05276490`:
```
ftep search -a SAMN05276490 --s2r
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
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(results[0].Records)
}
```
