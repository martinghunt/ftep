# Example URLs. Prefix them all with https://www.ebi.ac.uk/ena/portal/api/
# Get runs for a sample:
#   search?result=read_run&query=sample_accession=SAMEA2275804&fields=run_accession,sample_accession'
#
# Assembly:
#   search?result=assembly&query=accession=GCA_900137745&format=json&fields=accession,sample_accession,run_accession,scientific_name
# Or could do with filiereport, but is easier to use search so is same as
# for reads
#   filereport?result=assembly&accession=GCA_900137745&fields=accession
BASE_PORTAL_URL = "https://www.ebi.ac.uk/ena/portal/api/"

URL_SEARCH_DATA = {
    "assembly": {
        "main_type": "search",
        "result": "assembly",
    },
    "sample": {
        "main_type": "search",
        "result": "read_run",
    },
    "run": {
        "main_type": "search",
        "result": "read_run",
    },
}


ASSEMBLY_SMALL = ["accession", "sample_accession", "run_accession"]
ASSEMBLY_DEFAULT = ASSEMBLY_SMALL + ["scientific_name", "tax_id"]

RUN_SMALL = [
    "study_accession",
    "secondary_study_accession",
    "sample_accession",
    "secondary_sample_accession",
    "run_accession",
]
RUN_DEFAULT = RUN_SMALL + ["instrument_platform", "library_layout", "fastq_ftp"]
RUN_BIG = RUN_DEFAULT + [
    "center_name",
    "broker_name",
    "read_count",
    "base_count",
    "collection_date",
]

FIELD_PRESETS = {
    "assembly": {
        "SMALL": ASSEMBLY_SMALL,
        "DEFAULT": ASSEMBLY_DEFAULT,
        "BIG": ASSEMBLY_DEFAULT,
    },
    "sample": {
        "SMALL": RUN_SMALL,
        "DEFAULT": RUN_DEFAULT,
        "BIG": RUN_BIG,
    },
    "run": {
        "SMALL": RUN_SMALL,
        "DEFAULT": RUN_DEFAULT,
        "BIG": RUN_BIG,
    },
}
