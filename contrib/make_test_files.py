"""
This script creates versions of the source files with a limited number of
lines, allowing us to manually test the process quicker.

Usage:
    python contrib/make_test_files.py SOURCE_DIRECTORY TARGET_DIRECTORY

Example:
    pyhton contrib/make_test_files.py data/ data/mini/
"""

from functools import partial
from multiprocessing import Pool, cpu_count
from pathlib import Path
from sys import argv, exit
from zipfile import ZipFile


MAX_LINES = 10_000


def lines_from(path):
    with ZipFile(path) as archive:
        with archive.open(path.stem) as reader:
            for count, line in enumerate(reader):
                if count >= MAX_LINES:
                    break
                yield line


def minify(path, target_directory):
    minified = target_directory / path.name
    with ZipFile(minified, mode="w") as archive:
        with archive.open(minified.stem, mode="w") as writer:
            for line in lines_from(path):
                writer.write(line)


def load_args():
    try:
        *_, source, target = argv
    except ValueError:
        print(__doc__)
        exit(1)

    source, target = Path(source), Path(target)
    if not source.exists():
        print(f"ERROR: Source directory {source} does not exist.")
        print(__doc__)
        exit(1)

    files = tuple(source.glob("*.zip"))
    if not files:
        print(f"ERROR: Source directory {source} has no ZIP files.")
        print(__doc__)
        exit(1)

    target.mkdir(exist_ok=True, parents=True)
    return files, target


def status(total, count=0):
    print(f"Created: {count} of {total}", end="\r")


def main():
    files, target = load_args()
    total = len(files)
    create = partial(minify, target_directory=target)

    status(total)
    with Pool(cpu_count()) as pool:
        for count, _ in enumerate(pool.imap_unordered(create, files), 1):
            status(total, count)


if __name__ == "__main__":
    main()
