from ftep import filereport


def run(options):
    filereport.search(
        accession=options.accession,
        acc_file=options.acc_file,
        outformat=options.outfmt
    )
