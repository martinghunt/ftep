#!/usr/bin/env python3

import argparse
import logging

import ftep


def main(args=None):
    parser = argparse.ArgumentParser(
        prog="ftep",
        usage="ftep <command> <options>",
        description="ftep: query the ena",
    )
    parser.add_argument("--version", action="version", version=ftep.__version__)

    subparsers = parser.add_subparsers(title="Available commands", help="", metavar="")

    # ----------- general options common to all tasks ------------------------
    common_parser = argparse.ArgumentParser(add_help=False)
    common_parser.add_argument(
        "--debug",
        action="store_true",
        help="More verbose logging, and less file cleaning",
    )

    # ---------------------------- filereport ---------------------------------
    subparser_filereport = subparsers.add_parser(
        "filereport",
        parents=[common_parser],
        help="General filereport search. Try to guess from format of accession",
        usage="ftep filereport [options]",
        description="General filereport search. Try to guess from format of accession",
    )
    filereport_acc_group = subparser_filereport.add_mutually_exclusive_group(required=True)
    filereport_acc_group.add_argument(
        "-a", "--accession",
        help="Accession to search for",
    )
    filereport_acc_group.add_argument(
        "-f", "--acc_file",
        help="File of accessions to search for, one per line",
        metavar="FILENAME",
    )
    subparser_filereport.add_argument(
        "--outfmt",
        choices=["json", "tsv"],
        help="Output format json or tsv [%(default)s]",
        default="tsv",
    )
    subparser_filereport.set_defaults(func=ftep.tasks.filereport.run)


    args = parser.parse_args()
    if not hasattr(args, "func"):
        parser.print_help()
        return

    logging.basicConfig(
        format="[%(asctime)s ftep %(levelname)s] %(message)s",
        datefmt="%Y-%m-%dT%H:%M:%S%z",
    )
    log = logging.getLogger()
    if args.debug:
        log.setLevel(logging.DEBUG)
    else:
        log.setLevel(logging.INFO)

    if hasattr(args, "func"):
        args.func(args)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
