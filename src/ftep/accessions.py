import re

REGEXES = {
    re.compile(r"""(?P<acc>GCA_[0-9]+)(\.[0-9])*$"""): "assembly",
}


def identify_accession(accession):
    for r, acc_type in REGEXES.items():
        match = r.match(accession)
        if match is not None:
            if match.group("acc") is not None:
                return  match.group("acc"), acc_type
    return None, None

