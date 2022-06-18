# Running this script requires DATA_URL and POSTGRES_URI envvars. `DATA_URL`
# should be a `.tar` file with the downloaded files (no directories).

set -e

echo '==> Starting a Screen session named minha-receita…'
screen -r minha-receita

echo "==> Downloading $DATA_URL…"
mkdir data
curl -L $DATA_URL > data/data.tar

echo "==> Unarchiving $DATA_URL…"
tar -xvf data/data.tar
rm data.tar

echo '==> Checking files…'
minha-receita check

echo '==> Starting the ETL…'
minha-receita transform --clean-up
