import json
import logging
import requests


def request(url, data, to_json=True):
    logging.debug(f"query url '{url}' with {data}")

    try:
        r = requests.get(url, data)
    except:
        logging.debug(r.url)
        raise Exception("Error reqesting data, {url} {data}")

    logging.debug(f"request url: {r.url}; status ok: {r.status_code == requests.codes.ok}" )

    if r.status_code != requests.codes.ok:
        raise Exception(f"Error requesting data. Error code={r.status_code}. {r.url}")

    if not to_json:
        return  r.text

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
