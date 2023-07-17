# ftep

Finding things in the ENA portal.

Currently supported: run, sample, assembly accessions.


## Install

Clone this repo, and run
```
python3 -m pip install .
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


