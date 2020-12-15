package transform

import "github.com/frictionlessdata/tableschema-go/schema"

// Schema extends `schema.Schema` adding a `Headers()` method.
type Schema schema.Schema

// Headers returns an array with strings with the names of the fields.
func (s *Schema) Headers() []string {
	var h []string
	for _, f := range s.Fields {
		h = append(h, f.Name)
	}
	return h
}

// CNAESchema follows Frictionless Data's table schema.
var CNAESchema = Schema{
	Fields: []schema.Field{
		{
			Name:        "cnpj",
			Type:        schema.StringType,
			Constraints: schema.Constraints{Required: true, Pattern: "\\d{14}"},
		},
		{
			Name:        "cnae",
			Type:        schema.IntegerType,
			Constraints: schema.Constraints{Required: true},
		},
	},
}

// PartnerSchema follows Frictionless Data's table schema.
var PartnerSchema = Schema{
	Fields: []schema.Field{
		{
			Name:        "cnpj",
			Type:        schema.StringType,
			Constraints: schema.Constraints{Required: true, Pattern: "\\d{14}"},
		},
		{
			Name:        "identificador_de_socio",
			Type:        schema.IntegerType,
			Constraints: schema.Constraints{Enum: []interface{}{1, 2, 3}},
		},
		{
			Name: "nome_socio",
			Type: schema.StringType,
		},
		{
			Name:          "cnpj_cpf_do_socio",
			Type:          schema.StringType,
			MissingValues: map[string]struct{}{"000": struct{}{}},
		},
		{
			Name: "codigo_qualificacao_socio",
			Type: schema.IntegerType,
		},
		{
			Name: "percentual_capital_social",
			Type: schema.NumberType,
		},
		{
			Name: "data_entrada_sociedade",
			Type: schema.DateType,
		},
		{
			Name:          "cpf_representante_legal",
			Type:          schema.StringType,
			MissingValues: map[string]struct{}{"": struct{}{}},
		},
		{
			Name: "nome_representante_legal",
			Type: schema.StringType,
		},
		{
			Name: "codigo_qualificacao_representante_legal",
			Type: schema.StringType,
		},
	},
}

// CompanySchema follows Frictionless Data's table schema.
var CompanySchema = Schema{
	Fields: []schema.Field{
		{
			Name:        "cnpj",
			Type:        schema.StringType,
			Constraints: schema.Constraints{Required: true, Pattern: "\\d{14}"},
		},
		{
			Name:        "identificador_matriz_filial",
			Type:        schema.IntegerType,
			Constraints: schema.Constraints{Enum: []interface{}{1, 2}},
		},
		{
			Name: "razao_social",
			Type: schema.StringType,
		},
		{
			Name: "nome_fantasia",
			Type: schema.StringType,
		},
		{
			Name:        "situacao_cadastral",
			Type:        schema.IntegerType,
			Constraints: schema.Constraints{Enum: []interface{}{1, 2, 3, 4, 8}},
		},
		{
			Name: "data_situacao_cadastral",
			Type: schema.DateType,
		},
		{
			Name: "motivo_situacao_cadastral",
			Type: schema.StringType,
		},
		{
			Name: "nome_cidade_exterior",
			Type: schema.StringType,
		},
		{
			Name: "codigo_natureza_juridica",
			Type: schema.IntegerType,
		},
		{
			Name: "data_inicio_atividade",
			Type: schema.DateType,
		},
		{
			Name: "cnae_fiscal",
			Type: schema.IntegerType,
		},
		{
			Name: "descricao_tipo_logradouro",
			Type: schema.StringType,
		},
		{
			Name: "logradouro",
			Type: schema.StringType,
		},
		{
			Name: "numero",
			Type: schema.StringType,
		},
		{
			Name: "complemento",
			Type: schema.StringType,
		},
		{
			Name: "bairro",
			Type: schema.StringType,
		},
		{
			Name: "cep",
			Type: schema.StringType,
		},
		{
			Name:        "uf",
			Type:        schema.StringType,
			Constraints: schema.Constraints{MinLength: 2, MaxLength: 2}},
		{
			Name: "codigo_municipio",
			Type: schema.IntegerType,
		},
		{
			Name: "municipio",
			Type: schema.StringType,
		},
		{
			Name: "ddd_telefone1",
			Type: schema.StringType,
		},
		{
			Name: "ddd_telefone2",
			Type: schema.StringType,
		},
		{
			Name: "ddd_fax",
			Type: schema.StringType,
		},
		{
			Name: "qualificacao_do_responsavel",
			Type: schema.IntegerType,
		},
		{
			Name: "capital_social",
			Type: schema.NumberType,
		},
		{
			Name:        "porte",
			Type:        schema.IntegerType,
			Constraints: schema.Constraints{Enum: []interface{}{0, 1, 3, 5}},
		},
		{
			Name:        "opcao_pelo_simples",
			Type:        schema.BooleanType,
			TrueValues:  []string{"5", "7"},
			FalseValues: []string{"", "0", "6", "8"},
			Constraints: schema.Constraints{Enum: []interface{}{"", 0, 5, 6, 7, 8}},
		},
		{
			Name: "data_opcao_pelo_simples",
			Type: schema.DateType,
		},
		{
			Name: "data_exclusao_do_simples",
			Type: schema.DateType,
		},
		{
			Name:        "opcao_pelo_mei",
			Type:        schema.BooleanType,
			TrueValues:  []string{"S"},
			FalseValues: []string{"N"},
			Constraints: schema.Constraints{Enum: []interface{}{"S", "N", ""}},
		},
		{
			Name: "situacao_especial",
			Type: schema.StringType,
		},
		{
			Name: "data_situacao_especial",
			Type: schema.StringType,
		},
	},
	PrimaryKeys: []string{"cnpj"},
}
