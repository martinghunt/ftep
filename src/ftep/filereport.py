import json
import logging
import re
import requests
import sys

from ftep import accessions


# Example URLs. Prefix them all with https://www.ebi.ac.uk/ena/portal/api/
# Get runs for a sample:
#   search?result=read_run&query=sample_accession=SAMEA2275804&fields=run_accession,sample_accession'
#
# Get assembly its samples:
#   filereport?result=assembly&accession=GCA_900137745&fields=accession

BASE_URL = "https://www.ebi.ac.uk/ena/portal/api/"

URL_DATA = {
    "assembly": {
        "main_type": "filereport",
        "result": "assembly",
        "fields": ["accession", "sample_accession", "run_accession", "scientific_name"],
    },
    "sample": {
        "main_type": "search",
        "result": "read_run",
        "fields": ["study_accession", "secondary_study_accession", "sample_accession", "secondary_sample_accession", "run_accession"],
    },
}


def search_key_value(query_type, accession):
    if query_type == "assembly":
        return "accession", accession
    elif query_type =="sample":
        return "query", f"sample_accession={accession}"
    else:
        raise NotImplementedError(f"query_type")


def filereport(accession, accession_type, fields=None):
    url_data = URL_DATA[accession_type]
    url = f"{BASE_URL}{url_data['main_type']}?"
    search_key, search_val = search_key_value(accession_type, accession)

    data = {
        "result": url_data["result"],
        search_key: search_val,
        "format": "json"
    }
    if fields is None:
        data["fields"] = ",".join(url_data["fields"])
    else:
        data["fields"] = ",".join(fields)

    logging.debug(f"query url '{url}' with {data}")

    try:
        r = requests.get(url, data)
    except:
        logging.debug(r.url)
        raise Exception("Error querying ENA accession={accession} {r.url}")

    logging.debug(f"request url: {r.url}; status ok: {r.status_code == requests.codes.ok}" )

    if r.status_code != requests.codes.ok:
        raise Exception(f"Error requesting data. Error code={r.status_code}. {r.url}")

    try:
        results = json.loads(r.text)
    except:
        raise Exception(f"Error parsing json from query:\n{r.text}")
    logging.debug(json.dumps(results, indent=2))

    for d in results:
        for k, v  in d.items():
            if v == "":
                d[k] = None
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
            new_results = filereport(fixed_accession, result_type, fields=fields)
        except:
            logging.warning(f"Error getting data for accession {accession}. Skipping")
            continue

        if len(new_results) == 0:
            logging.warning(f"No results returned for accession {accession}. Skipping")
            continue

        logging.debug(f"results for {accession}: {new_results}")
        if outformat == "tsv":
            if columns is None:
                columns = list(new_results[0].keys())
                print("input_accession", *columns, sep="\t")

            for result in new_results:
                print(accession, *[replace_none.get(result[x], result[x]) for x in columns], sep="\t")
        else:
            results[accession] = new_results

    if outformat == "json":
        print(json.dumps(results, indent=2))
