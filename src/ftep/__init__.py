from pkg_resources import get_distribution

try:
    __version__ = get_distribution("ftep").version
except:
    __version__ = "local"


__all__ = [
    "accessions",
    "constants",
    "search",
    "tasks",
    "utils",
]

from ftep import *
