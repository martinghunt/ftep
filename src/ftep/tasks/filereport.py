from ftep import filereport


def run(options):
    fields = None if options.columns is None else options.columns.split(",")
    filereport.search(
        accession=options.accession,
        acc_file=options.acc_file,
        outformat=options.outfmt,
        fields=fields
    )
