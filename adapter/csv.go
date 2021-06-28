package adapter

import (
	"encoding/csv"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz"
)

const separator = ';'

func headersFor(a *Adapter) []string {
	switch a.kind {
	case company:
		return []string{
			"cnpj",
			"razao_social",
			"natureza_juridica",
			"qualificacao_do_responsavel",
			"capital_social",
			"porte",
			"ente_federativo_responsavel",
		}
	case facility:
		return []string{
			"cnpj",
			"cnpj_ordem",
			"cnpj_digito_verificador",
			"identificador_matriz_filial",
			"nome_fantasia",
			"situacao_cadastral",
			"data_situacao_cadastral",
			"motivo_situacao_cadastral",
			"nome_cidade_exterior",
			"pais",
			"data_inicio_atividade",
			"cnae_principal",
			"cnae_secundaria",
			"tipo_logradouro",
			"logradouro",
			"numero",
			"complemento",
			"bairro",
			"cep",
			"uf",
			"municipio",
			"ddd1",
			"telefone1",
			"ddd2",
			"telefone2",
			"ddd_fax",
			"fax",
			"email",
			"situacao_especial",
			"data_situacao_especial",
		}
	case partner:
		return []string{
			"cnpj",
			"identificador",
			"nome_razao_social",
			"cpf_cnpj",
			"qualificacao",
			"data_entrada",
			"pais",
			"cpf_representante_legal",
			"nome_representante_legal",
			"qualificacao_representante_legal",
			"faixa_etaria",
		}
	}
	return []string{}
}

func csvPathFor(a *Adapter) string {
	var n string
	switch a.kind {
	case company:
		n = "empresa.csv.xz"
	case facility:
		n = "estabelecimento.csv.xz"
	case partner:
		n = "socio.csv.xz"
	}

	return filepath.Join(a.dir, n)
}

func createCsvFor(a *Adapter) error {
	p := csvPathFor(a)
	if err := os.RemoveAll(p); err != nil {
		return err
	}

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	x, err := xz.NewWriter(f)
	if err != nil {
		return err
	}
	defer x.Close()

	w := csv.NewWriter(x)

	if err := w.Write(headersFor(a)); err != nil {
		return err
	}
	w.Flush()

	return nil
}
