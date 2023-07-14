import re

# see https://ena-docs.readthedocs.io/en/latest/submit/general-guide/accessions.html
REGEXES = {
    re.compile(r"""(?P<acc>GCA_[0-9]{9})(\.[0-9]*)*$"""): "assembly",
    re.compile(r"""(?P<acc>SAM(E|D|N)[A-Z]?[0-9]+)$"""): "sample",
    re.compile(r"""(?P<acc>(E|D|S)RS[0-9]{6,})$"""): "sample",
}


def identify_accession(accession):
    for r, acc_type in REGEXES.items():
        match = r.match(accession)
        if match is not None:
            if match.group("acc") is not None:
                return  match.group("acc"), acc_type
    return None, None

