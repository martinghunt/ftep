import json
import logging
import re
import requests
import sys

from ftep import accessions

DEFAULT_FIELDS = {
    "assembly": ["accession", "sample_accession", "run_accession", "scientific_name"],
}


def filereport(accession, result, fields=None):
    url = "https://www.ebi.ac.uk/ena/portal/api/filereport?"
    data = {
        "accession": accession,
        "result": result,
    }
    if fields is None and result in DEFAULT_FIELDS:
        data["fields"] = DEFAULT_FIELDS[result]
    else:
        data["fields"] = fields

    try:
        r = requests.get(url, data)
    except:
        raise Exception("Error querying ENA accession={accession} {r.url}")

    if r.status_code != requests.codes.ok:
        raise Exception(f"Error requesting data. Error code={r.status_code}. {r.url}")

    lines = r.text.rstrip("\n").split("\n")
    field_names = lines[0].split("\t")
    results = []
    for line in lines:
        field_vals = []
        for field in lines[1].split("\t"):
            field_vals.append(None if field == "" else field)
        if len(field_names) != len(field_vals):
            raise Exception("Mismatch number of column names and number of columns\n" + "\n".join(lines))
        try:
            results = dict(zip(field_names, field_vals))
        except:
            raise Exception("Error making dict from column names and results\n{field_names}\n{fields_vals}")

    return results



def search(accession=None, acc_file=None, fields=None, outformat="tsv"):
    to_search = []
    if accession is not None:
        to_search.append(accession)
    if acc_file is not None:
        with open(acc_file) as f:
            to_search.extend([x.rstrip() for x in f])

    to_search = {x: accessions.identify_accession(x) for x in to_search}
    result_types = list(set(x[1] for x in to_search.values()))
    if len(result_types) > 1 or result_types == {None}:
        for accession, (fixed_accession, res_type) in to_search.items():
            print(accession, res_type, sep="\t", file=sys.stderr)
            raise Exception(f"Error getting result types from accessions. See above output")

    results = {}
    columns = None
    replace_none = {None: "."}

    for accession, (fixed_accession, result_type) in to_search.items():
        assert result_type is not None

        try:
            result = filereport(fixed_accession, result_type, fields=fields)
        except:
            logging.warning(f"Error getting data for accession {accession}. Skipping")
            continue

        if outformat == "tsv":
            if columns is None:
                columns = list(result.keys())
                print("original_accession", *columns, sep="\t")
            print(accession, *[replace_none.get(result[x], result[x]) for x in columns], sep="\t")
        else:
            results[accession] = result

    if outformat == "json":
        print(json.dumps(results, indent=2))
