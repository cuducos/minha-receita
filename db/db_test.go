package db

import "testing"

const expected = `{"cnpj":"19131243000197","identificador_matriz_filial":0,"razao_social":"","nome_fantasia":"","situacao_cadastral":0,"data_situacao_cadastral":"0001-01-01T00:00:00Z","motivo_situacao_cadastral":0,"nome_cidade_exterior":"","codigo_natureza_juridica":0,"data_inicio_atividade":"0001-01-01T00:00:00Z","cnae_fiscal":0,"cnae_fiscal_descricao":"","descricao_tipo_logradouro":"","logradouro":"","numero":"","complemento":"","bairro":"","cep":0,"uf":"","codigo_municipio":0,"municipio":"","ddd_telefone_1":"","ddd_telefone_2":"","ddd_fax":"","qualificacao_do_responsavel":0,"capital_social":0,"porte":0,"opcao_pelo_simples":false,"data_opcao_pelo_simples":"","data_exclusao_do_simples":"","opcao_pelo_mei":false,"situacao_especial":"","data_situacao_especial":"","qsa":null,"cnaes_secundarias":null}`

func TestCompany(t *testing.T) {
	c := Company{Cnpj: "19131243000197"}

	if j, _ := c.JSON(); j != expected {
		t.Errorf("\nExpected JSON to be:\n\t%s\nGot:\n\t%s", expected, j)
	}

	if s := c.String(); s != "19.131.243/0001-97" {
		t.Errorf("Expected company to be 19.131.243/0001-97, but got %s", s)
	}
}
