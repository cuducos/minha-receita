package transform

import "testing"

func TestSchema(t *testing.T) {
	for _, c := range []struct {
		schema  Schema
		headers []string
	}{
		{schema: CompanySchema,
			headers: []string{
				"cnpj",
				"identificador_matriz_filial",
				"razao_social",
				"nome_fantasia",
				"situacao_cadastral",
				"data_situacao_cadastral",
				"motivo_situacao_cadastral",
				"nome_cidade_exterior",
				"codigo_natureza_juridica",
				"data_inicio_atividade",
				"cnae_fiscal",
				"descricao_tipo_logradouro",
				"logradouro",
				"numero",
				"complemento",
				"bairro",
				"cep",
				"uf",
				"codigo_municipio",
				"municipio",
				"ddd_telefone1",
				"ddd_telefone2",
				"ddd_fax",
				"qualificacao_do_responsavel",
				"capital_social",
				"porte",
				"opcao_pelo_simples",
				"data_opcao_pelo_simples",
				"data_exclusao_do_simples",
				"opcao_pelo_mei",
				"situacao_especial",
				"data_situacao_especial",
			},
		}, {
			schema: PartnerSchema,
			headers: []string{
				"cnpj",
				"identificador_de_socio",
				"nome_socio",
				"cnpj_cpf_do_socio",
				"codigo_qualificacao_socio",
				"percentual_capital_social",
				"data_entrada_sociedade",
				"cpf_representante_legal",
				"nome_representante_legal",
				"codigo_qualificacao_representante_legal",
			},
		}, {
			schema:  CNAESchema,
			headers: []string{"cnpj", "cnae"},
		},
	} {
		got := c.schema.Headers()
		if len(got) != len(c.headers) {
			t.Errorf("Expected headers to have %d items, got %d", len(c.headers), len(got))
			continue
		}
		for i, g := range got {
			if g != c.headers[i] {
				t.Errorf("Expected item #%d to be %s, got %s", i, c.headers[i], g)
			}
		}
	}
}
