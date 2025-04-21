package db

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoDB(t *testing.T) {
	id := "33683111000280"
	json := `{"cnpj":"33683111000280","identificador_matriz_filial":null,"descricao_identificador_matriz_filial":null,"nome_fantasia":"Minh","situacao_cadastral":null,"descricao_situacao_cadastral":null,"data_situacao_cadastral":null,"motivo_situacao_cadastral":null,"descricao_motivo_situacao_cadastral":null,"nome_cidade_no_exterior":"","codigo_pais":null,"pais":null,"data_inicio_atividade":null,"cnae_fiscal":null,"cnae_fiscal_descricao":null,"descricao_tipo_de_logradouro":"","logradouro":"","numero":"","complemento":"","bairro":"","cep":"","uf":"","codigo_municipio":null,"codigo_municipio_ibge":null,"municipio":null,"ddd_telefone_1":"","ddd_telefone_2":"","ddd_fax":"","email":null,"situacao_especial":"","data_situacao_especial":null,"opcao_pelo_simples":null,"data_opcao_pelo_simples":null,"data_exclusao_do_simples":null,"opcao_pelo_mei":null,"data_opcao_pelo_mei":null,"data_exclusao_do_mei":null,"razao_social":"","codigo_natureza_juridica":null,"natureza_juridica":null,"qualificacao_do_responsavel":null,"capital_social":null,"codigo_porte":null,"porte":null,"ente_federativo_responsavel":"","descricao_porte":"","qsa":[{"identificador_de_socio":null,"nome_socio":"OSMAR QUIRINO DA SILVA","cnpj_cpf_do_socio":"","codigo_qualificacao_socio":null,"qualificacao_socio":"Diretor","data_entrada_sociedade":null,"codigo_pais":null,"pais":null,"cpf_representante_legal":"","nome_representante_legal":"","codigo_qualificacao_representante_legal":null,"qualificacao_representante_legal":null,"codigo_faixa_etaria":null,"faixa_etaria":"Entre 61 a 70 anos"},{"identificador_de_socio":null,"nome_socio":"ALEXANDRE GONCALVES DE AMORIM","cnpj_cpf_do_socio":"","codigo_qualificacao_socio":null,"qualificacao_socio":"Presidente","data_entrada_sociedade":null,"codigo_pais":null,"pais":null,"cpf_representante_legal":"","nome_representante_legal":"","codigo_qualificacao_representante_legal":null,"qualificacao_representante_legal":null,"codigo_faixa_etaria":null,"faixa_etaria":"Entre 51 a 60 anos"}],"cnaes_secundarios":null,"regime_tributario":null}`

	u := os.Getenv("TEST_MONGODB_URL")
	if u == "" {
		t.Errorf("expected a mongodb uri at TEST_MONGODB_URL, found nothing")
		return
	}
	db, err := NewMongoDB(u)
	if err != nil {
		t.Errorf("expected no error connecting to mongodb, got %s", err)
		return
	}
	if err := db.Drop(); err != nil {
		t.Errorf("expected no error dropping the table, got %s", err)
	}
	defer func() {
		if err := db.Drop(); err != nil {
			t.Errorf("expected no error dropping the table, got %s", err)
		}
		db.Close()
	}()

	if err := db.Create(); err != nil {
		t.Errorf("expected no error creating the table, got %s", err)
	}

	if err := db.CreateCompanies([][]string{{id, json}}); err != nil {
		t.Errorf("expected no error saving a company, got %s", err)
	}
	if err := db.CreateCompanies([][]string{{id, json}}); err != nil {
		t.Errorf("expected no error saving a duplicated company, got %s", err)
	}
	if err := db.PostLoad(); err != nil {
		t.Errorf("expected no error post load, got %s", err)
	}
	got, err := db.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	if got != json {
		t.Errorf("expected json to be %s, got %s", json, got)
	}
	got, err = db.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	if got != json {
		t.Errorf("expected json to be %s, got %s", json, got)
	}
	if err := db.MetaSave("answer", "42"); err != nil {
		t.Errorf("expected no error writing to the metadata table, got %s", err)
	}
	metadata, err := db.MetaRead("answer")
	if err != nil {
		t.Errorf("expected no error getting metadata, got %s", err)
	}
	if metadata != "42" {
		t.Errorf("expected 42 as the answer, got %s", metadata)
	}
	if err := db.MetaSave("answer", "forty-two"); err != nil {
		t.Errorf("expected no error re-writing to the metadata table, got %s", err)
	}
	metadata2, err := db.MetaRead("answer")
	if err != nil {
		t.Errorf("expected no error getting metadata for the second time, got %s", err)
	}
	if metadata2 != "forty-two" {
		t.Errorf("expected foruty-two as the answer, got %s", metadata2)
	}
	if err := db.ExtraIndexes([]string{"teste.index1"}); err != nil {
		t.Errorf("error creating new index, got %s", err) // preciso de ajuda ao melhorar a mensagem de erro.
	}
	c := db.db.Collection(companyTableName)
	cur, err := c.Indexes().List(db.ctx)
	if err != nil {
		t.Errorf("expected no errors checking index list, got %s", err)
	}
	defer cur.Close(db.ctx)
	idxs := make(map[string]bool)
	for cur.Next(db.ctx) {
		var idx bson.M
		if err := cur.Decode(&idx); err != nil {
			t.Errorf("error decoding index: %s", err) // mais uma ajuda
		}
		if n, ok := idx["name"].(string); ok {
			idxs[n] = true
		}
	}
	r := make(map[string]bool)
	for _, index := range []string{"index1"} {
		r[index] = idxs[index]
	}
	if len(r) == 0 {
		t.Errorf("index not found, got %s", err) // mais uma ajuda
	}
}
