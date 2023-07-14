import pytest

from ftep import accessions


def test_identify_accession():
    f = accessions.identify_accession
    assert f("GCA_123456") == ("GCA_123456", "assembly")
    assert f("GCA_123456.1") == ("GCA_123456","assembly")
    assert f("G123456.1") == (None, None)
