from multiprocessing import Pool, cpu_count
from pathlib import Path
from sys import argv
from zipfile import ZipFile


def unzip(path: Path) -> Path:
    dir = path.parent / "csv"
    dir.mkdir(exist_ok=True)
    print(f"Unzipping {path.name}…")
    with ZipFile(path, "r") as source:
        source.extract(path.stem, path=path.parent / "csv")
    return Path(dir / path.stem).rename(dir / f"{path.stem}.csv")


def run() -> None:
    dir = Path(argv[1])
    if not dir.exists():
        raise RuntimeError(f"Directory {dir} does not exist.")

    with Pool(cpu_count()) as pool:
        files = tuple(dir.glob("*.zip"))
        tasks = pool.imap_unordered(unzip, files)
        for count, path in enumerate(tasks, 1):
            print(f"[{count} of {len(files)}] {path.name} done!")


if __name__ == "__main__":
    run()
