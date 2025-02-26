import json
import logging
import re
import sys

from ftep import accessions, constants, utils


def search_key_value(query_type, accession):
    if query_type == "assembly":
        return (
            "query",
            f"accession={accession}",
        )
    elif query_type == "sample":
        return (
            "query",
            f"sample_accession={accession} OR secondary_sample_accession={accession}",
        )
    elif query_type == "run":
        return "query", f"run_accession={accession}"
    elif query_type == "experiment":
        return "query", f"experiment_accession={accession}"
    else:
        raise NotImplementedError(f"query_type")


def ena_query(accession, accession_type, fields=None, sample2run=False):
    search_key, search_val = search_key_value(accession_type, accession)
    if sample2run and accession_type == "sample":
        logging.debug("Getting run data instead of sample data")
        accession_type = "run"
    url_data = constants.URL_SEARCH_DATA[accession_type]
    url = f"{constants.BASE_PORTAL_URL}{url_data['main_type']}?"

    data = {"result": url_data["result"], search_key: search_val, "format": "json"}
    field_presets = constants.FIELD_PRESETS[accession_type]
    if fields is None:
        fields = ["DEFAULT"]

    if fields[0] in field_presets:
        fields = field_presets[fields[0]]

    data["fields"] = ",".join(fields)

    results = utils.request(url, data)
    return fields, results


def search(
    accession=None, acc_file=None, fields=None, outformat="tsv", sample2run=False
):
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
            raise Exception(
                f"Error getting result types from accessions. See above output"
            )

    results = {}
    columns = None
    replace_none = {None: "."}

    for accession, (fixed_accession, result_type) in to_search.items():
        logging.debug(f"Search for {accession}")
        assert result_type is not None

        try:
            new_fields, new_results = ena_query(
                fixed_accession,
                result_type,
                fields=fields,
                sample2run=sample2run,
            )
        except:
            logging.warning(f"Error getting data for accession {accession}. Skipping")
            continue

        if len(new_results) == 0:
            logging.warning(f"No results returned for accession {accession}. Skipping")
            continue

        logging.debug(f"results for {accession}: {new_results}")
        if outformat == "tsv":
            if columns is None:
                if fields == ["ALL"]:
                    columns = sorted(list(new_results[0].keys()))
                else:
                    columns = new_fields
                print("input_accession", *columns, sep="\t")
            else:
                assert set(columns) == set(new_fields)

            for result in new_results:
                print(
                    accession,
                    *[replace_none.get(result[x], result[x]) for x in columns],
                    sep="\t",
                )
        else:
            results[accession] = new_results

    if outformat == "json":
        print(json.dumps(results, indent=2))

    return results


def get_allowed_fields(data_type):
    url = "https://www.ebi.ac.uk/ena/portal/api/searchFields?"
    results = utils.request(url, {"result": data_type}, to_json=False)
    print(results)
