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
        # "fields": ["accession", "sample_accession", "run_accession", "scientific_name"],
    },
    "sample": {
        "main_type": "search",
        "result": "read_run",
        # "fields": ["study_accession", "secondary_study_accession", "sample_accession", "secondary_sample_accession", "run_accession"],
    },
}


ASSEMBLY_SMALL = ["accession", "sample_accession", "run_accession"]
ASSEMBLY_DEFAULT = ASSEMBLY_SMALL + ["scientific_name", "tax_id"]

SAMPLE_SMALL = [
    "study_accession",
    "secondary_study_accession",
    "sample_accession",
    "secondary_sample_accession",
    "run_accession",
]
SAMPLE_DEFAULT = SAMPLE_SMALL + ["instrument_platform", "library_layout", "fastq_ftp"]
SAMPLE_BIG = SAMPLE_DEFAULT + [
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
        "SMALL": SAMPLE_SMALL,
        "DEFAULT": SAMPLE_DEFAULT,
        "BIG": SAMPLE_BIG,
    },
}
