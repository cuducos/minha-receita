rows pgimport --schema /mnt/schemas/empresa.csv /mnt/data/empresa.csv.gz $POSTGRES_URI empresa
rows pgimport --schema /mnt/schemas/socio.csv /mnt/data/socio.csv.gz $POSTGRES_URI socio
rows pgimport --schema /mnt/schemas/cnae-secundaria.csv /mnt/data/cnae-secundaria.csv.gz $POSTGRES_URI cnae_secundaria
psql $POSTGRES_URI -c "CREATE INDEX idx_empresa_cnpj ON empresa(cnpj);"
psql $POSTGRES_URI -c "CREATE INDEX idx_socio_cnpj ON socio(cnpj);"
psql $POSTGRES_URI -c "CREATE INDEX idx_cnae_secundaria_cnpj ON cnae_secundaria(cnpj);"
