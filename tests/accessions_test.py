import pytest

from ftep import accessions


def test_identify_accession():
    f = accessions.identify_accession
    assert f("GCA_123456789") == ("GCA_123456789", "assembly")
    assert f("GCA_123456789.1") == ("GCA_123456789","assembly")
    assert f("GCA_12345678.1") == (None, None) # too short
    assert f("G123456.1") == (None, None)

    assert f("SAMN123456") == ("SAMN123456", "sample")
    assert f("ERS123456") == ("ERS123456", "sample")
    assert f("ERS12345") == (None, None) # too short

    assert f("ERR123456") == ("ERR123456", "run")
    assert f("ERR12345") == (None, None) # too short
