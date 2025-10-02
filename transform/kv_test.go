package transform

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/cuducos/minha-receita/testutils"
	"github.com/dgraph-io/badger/v4"
)

var (
	taxRegimeRow      = []string{"2022", "BASE DO CNPJ", "CNPJ DA SCP", "FORMA DE TRIBUTAÇÃO", "42"}
	expectedTaxregime = []byte(`{"ano":2022,"cnpj_da_scp":"CNPJ DA SCP","forma_de_tributacao":"FORMA DE TRIBUTAÇÃO","quantidade_de_escrituracoes":42}`)
)

func TestBadgerStorageClose(t *testing.T) {
	kv, err := newBadgerStorage(t.TempDir(), false)
	if err != nil {
		t.Errorf("expected no error creating badger storage, got %s", err)
	}
	if err := kv.close(); err != nil {
		t.Errorf("expected no error closing badger storage, got %s", err)
	}
}

func TestNewItem(t *testing.T) {
	l, err := newLookups(testdata)
	if err != nil {
		t.Fatalf("could not create lookus: %s", err)
	}
	for _, tc := range []struct {
		kind      sourceType
		row       []string
		keyPrefix []byte
		value     []byte
	}{
		{partners, partnerCSVRow, []byte("p-BASE DO CNPJ"), toBytes(t, newTestPartner())},
		{base, baseCSVRow, []byte("b-BASE DO CNPJ"), toBytes(t, newTestBaseCNPJ())},
		{simpleTaxes, simpleTaxesCSVRow, []byte("st-BASE DO CNPJ"), toBytes(t, newTestTaxes())},
		{realProfit, taxRegimeRow, []byte("tr-BASEDOCNPJ"), expectedTaxregime},
		{presumedProfit, taxRegimeRow, []byte("tr-BASEDOCNPJ"), expectedTaxregime},
		{arbitratedProfit, taxRegimeRow, []byte("tr-BASEDOCNPJ"), expectedTaxregime},
		{noTaxes, taxRegimeRow, []byte("tr-BASEDOCNPJ"), expectedTaxregime},
	} {
		t.Run(string(tc.kind), func(t *testing.T) {
			got, err := newKVItem(tc.kind, &l, tc.row)
			if err != nil {
				t.Errorf("could not create key-value item: %s", err)
			}
			if !strings.HasPrefix(string(got.key), string(tc.keyPrefix)) {
				t.Errorf("expected item's key to start with %s, got %s", string(tc.keyPrefix), string(got.key))
			}
			if string(got.value) != string(tc.value) {
				t.Errorf("expected item's value to be %s, got %s", string(tc.value), string(got.value))
			}
			if string(got.kind) != string(tc.kind) {
				t.Errorf("expected item's kind to be %s, got %s", string(tc.kind), string(got.kind))
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("single values", func(t *testing.T) {
		l, err := newLookups(testdata)
		if err != nil {
			t.Fatalf("could not create lookups: %s", err)
		}
		kv, err := newBadgerStorage(t.TempDir(), false)
		if err != nil {
			t.Fatalf("could not create badger storage: %s", err)
		}
		defer func() {
			if err := kv.close(); err != nil {
				t.Errorf("expected no error closing key-value storage, got %s", err)
			}
		}()
		if err := kv.load(testdata, &l, 1024); err != nil {
			t.Errorf("expected no error loading data, got %s", err)
		}
		for _, tc := range []struct{ key, value string }{
			{"b-19131243", `{"codigo_porte":5,"porte":"DEMAIS","razao_social":"OPEN KNOWLEDGE BRASIL","codigo_natureza_juridica":3999,"natureza_juridica":null,"qualificacao_do_responsavel":16,"capital_social":0,"ente_federativo_responsavel":""}`},
			{"b-33683111", `{"codigo_porte":5,"porte":"DEMAIS","razao_social":"SERVICO FEDERAL DE PROCESSAMENTO DE DADOS (SERPRO)","codigo_natureza_juridica":2011,"natureza_juridica":"Empresa Pública","qualificacao_do_responsavel":16,"capital_social":1061004800,"ente_federativo_responsavel":""}`},
			{"st-33683111", `{"opcao_pelo_simples":true,"data_opcao_pelo_simples":"2014-01-01","data_exclusao_do_simples":null,"opcao_pelo_mei":false,"data_opcao_pelo_mei":null,"data_exclusao_do_mei":null}`},
		} {
			assertKeyValue(t, kv, tc.key, tc.value)
		}
	})

	t.Run("accumulative values", func(t *testing.T) {
		l, err := newLookups(testdata)
		if err != nil {
			t.Fatalf("could not create lookups: %s", err)
		}
		kv, err := newBadgerStorage(t.TempDir(), false)
		if err != nil {
			t.Fatalf("could not create badger storage: %s", err)
		}
		defer func() {
			if err := kv.close(); err != nil {
				t.Errorf("expected no error closing key-value storage, got %s", err)
			}
		}()
		if err := kv.load(testdata, &l, 1024); err != nil {
			t.Errorf("expected no error loading data, got %s", err)
		}
		for _, tc := range []struct {
			prefix string
			value  []string
		}{
			{"p-19131243", []string{`{"identificador_de_socio":2,"nome_socio":"FERNANDA CAMPAGNUCCI PEREIRA","cnpj_cpf_do_socio":"***690948**","codigo_qualificacao_socio":16,"qualificacao_socio":"Presidente","data_entrada_sociedade":"2019-10-25","codigo_pais":null,"pais":null,"cpf_representante_legal":"***000000**","nome_representante_legal":"","codigo_qualificacao_representante_legal":0,"qualificacao_representante_legal":null,"codigo_faixa_etaria":4,"faixa_etaria":"Entre 31 a 40 anos"}`}},
			{"p-33683111", []string{
				`{"identificador_de_socio":2,"nome_socio":"ANDRE DE CESERO","cnpj_cpf_do_socio":"***220050**","codigo_qualificacao_socio":10,"qualificacao_socio":"Diretor","data_entrada_sociedade":"2016-06-16","codigo_pais":null,"pais":null,"cpf_representante_legal":"***000000**","nome_representante_legal":"","codigo_qualificacao_representante_legal":0,"qualificacao_representante_legal":null,"codigo_faixa_etaria":6,"faixa_etaria":"Entre 51 a 60 anos"}`,
				`{"identificador_de_socio":2,"nome_socio":"ANTONIO DE PADUA FERREIRA PASSOS","cnpj_cpf_do_socio":"***595901**","codigo_qualificacao_socio":10,"qualificacao_socio":"Diretor","data_entrada_sociedade":"2016-12-08","codigo_pais":null,"pais":null,"cpf_representante_legal":"***000000**","nome_representante_legal":"","codigo_qualificacao_representante_legal":0,"qualificacao_representante_legal":null,"codigo_faixa_etaria":7,"faixa_etaria":"Entre 61 a 70 anos"}`,
				`{"identificador_de_socio":2,"nome_socio":"WILSON BIANCARDI COURY","cnpj_cpf_do_socio":"***414127**","codigo_qualificacao_socio":10,"qualificacao_socio":"Diretor","data_entrada_sociedade":"2019-06-18","codigo_pais":null,"pais":null,"cpf_representante_legal":"***000000**","nome_representante_legal":"","codigo_qualificacao_representante_legal":0,"qualificacao_representante_legal":null,"codigo_faixa_etaria":8,"faixa_etaria":"Entre 71 a 80 anos"}`,
				`{"identificador_de_socio":2,"nome_socio":"GILENO GURJAO BARRETO","cnpj_cpf_do_socio":"***099595**","codigo_qualificacao_socio":16,"qualificacao_socio":"Presidente","data_entrada_sociedade":"2020-02-03","codigo_pais":null,"pais":null,"cpf_representante_legal":"***000000**","nome_representante_legal":"","codigo_qualificacao_representante_legal":0,"qualificacao_representante_legal":null,"codigo_faixa_etaria":5,"faixa_etaria":"Entre 41 a 50 anos"}`,
				`{"identificador_de_socio":2,"nome_socio":"RICARDO CEZAR DE MOURA JUCA","cnpj_cpf_do_socio":"***989951**","codigo_qualificacao_socio":10,"qualificacao_socio":"Diretor","data_entrada_sociedade":"2020-05-12","codigo_pais":null,"pais":null,"cpf_representante_legal":"***000000**","nome_representante_legal":"","codigo_qualificacao_representante_legal":0,"qualificacao_representante_legal":null,"codigo_faixa_etaria":5,"faixa_etaria":"Entre 41 a 50 anos"}`,
				`{"identificador_de_socio":2,"nome_socio":"ANTONINO DOS SANTOS GUERRA NETO","cnpj_cpf_do_socio":"***073447**","codigo_qualificacao_socio":5,"qualificacao_socio":"Administrador","data_entrada_sociedade":"2019-02-11","codigo_pais":null,"pais":null,"cpf_representante_legal":"***000000**","nome_representante_legal":"","codigo_qualificacao_representante_legal":0,"qualificacao_representante_legal":null,"codigo_faixa_etaria":7,"faixa_etaria":"Entre 61 a 70 anos"}`,
			}},
		} {
			assertKeyValues(t, kv, tc.prefix, tc.value)
		}
	})
}

func TestEnrichCompany(t *testing.T) {
	l, err := newLookups(testdata)
	if err != nil {
		t.Fatalf("could not create lookups: %s", err)
	}
	kv, err := newBadgerStorage(t.TempDir(), false)
	if err != nil {
		t.Fatalf("could not create badger storage: %s", err)
	}
	defer func() {
		if err := kv.close(); err != nil {
			t.Errorf("error closing key-value storage: %s", err)
		}
	}()
	if err := kv.load(testdata, &l, 1024); err != nil {
		t.Errorf("expected no error loading data, got %s", err)
	}
	c := Company{CNPJ: "33683111000280"}
	if err := kv.enrichCompany(&c); err != nil {
		t.Errorf("expected no error enriching company, got %s", err)
	}
	if len(c.QuadroSocietario) != 6 {
		t.Errorf("expected 6 partners, got %d", len(c.QuadroSocietario))
	}
	if *c.CodigoPorte != 5 {
		t.Errorf("expected CodeSize to be 5, got %d", *c.CodigoPorte)
	}
	if *c.Porte != "DEMAIS" {
		t.Errorf("expected Porte to be DEMAIS, got %s", *c.Porte)
	}
	if c.RazaoSocial != "SERVICO FEDERAL DE PROCESSAMENTO DE DADOS (SERPRO)" {
		t.Errorf("expected RazaoSocial to be SERVICO FEDERAL DE PROCESSAMENTO DE DADOS (SERPRO), got %s", c.RazaoSocial)
	}
	if *c.CodigoNaturezaJuridica != 2011 {
		t.Errorf("expected CodigoNaturezaJuridica to be 2011, got %d", *c.CodigoNaturezaJuridica)
	}
	if *c.NaturezaJuridica != "Empresa Pública" {
		t.Errorf("expected NaturezaJuridica to be Empresa Pública, got %s", *c.NaturezaJuridica)
	}
	if *c.QualificacaoDoResponsavel != 16 {
		t.Errorf("expected QualificacaoDoResponsavel to be 16, got %d", *c.QualificacaoDoResponsavel)
	}
	if *c.CapitalSocial != float32(1061004800) {
		t.Errorf("expected CapitalSocial to be 1061004800, got %f", *c.CapitalSocial)
	}
	if c.EnteFederativoResponsavel != "" {
		t.Errorf("expected EnteFederativoResponsavel to be empty, got %s", c.EnteFederativoResponsavel)
	}
	if !*c.OpcaoPeloSimples {
		t.Errorf("expected OpcaoPeloSimples to be true, got %t", *c.OpcaoPeloSimples)
	}
	if !time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC).Equal(time.Time(*c.DataOpcaoPeloSimples)) {
		t.Errorf("expected DataOpcaoPeloSimples to be 2014-01-01, got %v", *c.DataOpcaoPeloSimples)
	}
	if c.DataExclusaoDoSimples != nil {
		t.Errorf("expected DataExclusaoDoSimples to be nil, got %v", *c.DataExclusaoDoSimples)
	}
	if *c.OpcaoPeloMEI {
		t.Errorf("expected OpcaoPeloSimples to be false, got %t", *c.OpcaoPeloMEI)
	}
	if c.DataOpcaoPeloMEI != nil {
		t.Errorf("expected DataOpcaoPeloMEI to be nil, got %v", *c.DataOpcaoPeloMEI)
	}
	if c.DataExclusaoDoMEI != nil {
		t.Errorf("expected DataExclusaoDoMEI to be nil, got %v", *c.DataExclusaoDoMEI)
	}
	if c.RegimeTributario == nil {
		t.Errorf("expected RegimeTributario not to be nil, got nil")
	}
	if len(c.RegimeTributario) != 1 {
		t.Errorf("expected RegimeTributario to have one record, got %d", len(c.RegimeTributario))
	}
	tr := TaxRegime{2018, nil, "LUCRO PRESUMIDO", 1}
	if c.RegimeTributario[0] != tr {
		t.Errorf("expected RegimeTributario to be TESTE, got %v", c.RegimeTributario[0])
	}
}

func assertKeyValue(t *testing.T, kv *badgerStorage, key, value string) {
	err := kv.db.View(func(tx *badger.Txn) error {
		i, err := tx.Get([]byte(key))
		if err != nil {
			return err
		}
		got, err := i.ValueCopy(nil)
		if err != nil {
			return err
		}
		if string(got) != value {
			t.Errorf("expected %s to be %s, got %s", string(key), value, string(got))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("could not read %s: %s", string(key), err)
	}
}

func assertKeyValues(t *testing.T, kv *badgerStorage, prefix string, value []string) {
	err := kv.db.View(func(tx *badger.Txn) error {
		var got []string
		it := tx.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.Valid() && bytes.HasPrefix(it.Item().Key(), []byte(prefix)); it.Next() {
			v, err := it.Item().ValueCopy(nil)
			if err != nil {
				return err
			}
			got = append(got, string(v))
		}
		testutils.AssertArraysHaveSameItems(t, value, got)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error searchinmg key-value storage for %s, got %s", string(prefix), err)
	}
}
