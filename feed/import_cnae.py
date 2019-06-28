from contextlib import contextmanager
from csv import DictWriter
from os import environ
from pathlib import Path
from re import sub
from subprocess import run
from tempfile import TemporaryDirectory
from typing import Iterator, NamedTuple, Optional

from openpyxl import load_workbook


# input data from: https://cnae.ibge.gov.br/classificacoes/download-concla.html
EXCEL_FILE = "/mnt/data/CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx"

Cnae = NamedTuple("CNAE", (("codigo", int), ("descricao", str)))
Schema = NamedTuple("Schema", (("field_name", str), ("field_type", str)))


def parse_code(code: str) -> Optional[int]:
    if not code:
        return

    cleaned = sub(r"\D", "", code)
    try:
        return int(cleaned)
    except ValueError:
        return


def cnaes(excel_file: Optional[str]) -> Iterator[Cnae]:
    excel_file = excel_file or EXCEL_FILE
    wb = load_workbook(excel_file)
    for row in wb.active.rows:
        code = parse_code(row[4].value)
        description = row[5].value

        if not all((code, description)):
            continue

        yield Cnae(code, description)


def csv_with_headers(path, rows, *fieldnames):
    with path.open("w") as fobj:
        writer = DictWriter(fobj, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(row._asdict() for row in rows)


@contextmanager
def cnaes_csv(excel_file=None):
    with TemporaryDirectory() as path:
        tmp = Path(path)
        data = tmp / "data.csv"
        schema = tmp / "schema.csv"

        csv_with_headers(data, cnaes(excel_file), "codigo", "descricao")
        csv_with_headers(
            schema,
            (Schema("codigo", "integer"), Schema("descricao", "text")),
            "field_name",
            "field_type",
        )
        yield data.as_posix(), schema.as_posix()


def main(excel_file=None):
    with cnaes_csv(excel_file) as files:
        data, schema = files
        postgres = environ.get("POSTGRES_URI")
        index = "CREATE INDEX idx_cnae_codigo ON cnae(codigo)"
        run(["rows", "pgimport", "--schema", schema, data, postgres, "cnae"])
        run(["psql", postgres, "-c", index])


if __name__ == "__main__":
    main()
