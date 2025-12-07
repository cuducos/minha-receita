package transformnext

import (
	"context"
	"testing"
	"time"
)

var dataSituacaoCadastral = date(time.Date(2004, 5, 22, 0, 0, 0, 0, time.UTC))

func TestMaskCPF(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		// MEI patterns (company name + CPF)
		{"João Silva 12345678901", "João Silva ***45678***"},
		{"Maria Santos ME 98765432109", "Maria Santos ME ***65432***"},
		{"JOSE DA SILVA 11122233344", "JOSE DA SILVA ***22233***"},
		{"COMERCIO DE ALIMENTOS LTDA 55566677788", "COMERCIO DE ALIMENTOS LTDA ***66677***"},
		// Edge cases with non-digit before CPF
		{"Empresa-12345678901", "Empresa-***45678***"},
		{"Nome 12345678901", "Nome ***45678***"},
		{"A12345678901", "A***45678***"},
		// Should NOT mask: 12 consecutive digits (not CPF pattern)
		{"Empresa123456789012", "Empresa123456789012"},
		{"000012345678901", "000012345678901"},
		// Should NOT mask: too short
		{"1234567890", "1234567890"},
		{"Short", "Short"},
		// Should NOT mask: non-digits in tail
		{"NomeEmpresa1234567890X", "NomeEmpresa1234567890X"},
		{"Empresa 1234567890a", "Empresa 1234567890a"},
		{"Test 123456-78901", "Test 123456-78901"},
		// Exactly 11 chars (all digits)
		{"12345678901", "***45678***"},
		// UTF-8 cases
		{"João José 12345678901", "João José ***45678***"},
		{"Quitanda São Miguel 99988877766", "Quitanda São Miguel ***88877***"},
		{"Café é Bom 12312312312", "Café é Bom ***12312***"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := maskCPF(tc.name)
			if got != tc.want {
				t.Errorf("expected masked %s to be %s, got %s", tc.name, tc.want, got)
			}
		})
	}
}

func TestNewCompany(t *testing.T) {
	row := []string{
		"33683111",             // 0 CNPJ Base
		"0002",                 // 1 CNPJ Ordem
		"80",                   // 2 CNPJ DV
		"2",                    // 3 Indentificador Matriz/Filial
		"REGIONAL BRASILIA-DF", // 4 Nome Fantasia
		"02",                   // 5 Situação Cadastral
		"20040522",             // 6 Data Situação Cadastral
		"00",                   // 7 Motivo Situação Cadastral
		"",                     // 8 Nome da cidade no exterior
		"",                     // 9 Pais
		"19670630",             // 10 Data de Início da Ativiade
		"6204000",              // 11 CNAE Fiscal
		"6201501,6202300,6203100,6209100,6311900", // 12 CNAEs Secundários
		"AVENIDA",      // 13 Tipo de Logradouro
		"L2 SGAN",      // 14 Logradouro
		"601",          // 15 Número
		"MODULO G",     // 16 Complemento
		"ASA NORTE",    // 17 Bairro
		"70836900",     // 18 CEP
		"DF",           // 19 UF
		"9701",         // 20 Município
		"",             // 21 DDD 1
		"",             // 22 Telefone 1
		"",             // 23 DDD 2
		"",             // 24 Telefone 2
		"",             // 25 DDD Fax
		"",             // 26 Fax
		"test@ser.pro", // 27 Email
		"",             // 28 Situação Especial
		"",             // 29 Data Situação Especial
	}
	kv, err := newBadger(t.TempDir(), false)
	if err != nil {
		t.Fatalf("expected no error creatinh kv, got %s", err)
	}
	srcs := sources()
	ctx := context.Background()
	for key, src := range srcs {
		if key == "est" {
			continue
		}
		if err := loadCSVs(ctx, "../testdata", src, nil, kv); err != nil {
			t.Fatalf("expected no error loading %s data, got %s", key, err)
		}
	}
	got, err := newCompany(srcs, kv, row)
	if err != nil {
		t.Fatalf("expected no error creating a company, got %s", err)
	}
	if got.CNPJ != "33683111000280" {
		t.Errorf("expected cnpj to be 33683111000280, got %s", got.CNPJ)
	}
	if *got.IdentificadorMatrizFilial != 2 {
		t.Errorf("expected IdentificadorMatrizFilial to be 2, got %v", got.IdentificadorMatrizFilial)
	}
	if *got.DescricaoMatrizFilial != "FILIAL" {
		t.Errorf("expected DescricaoMatrizFilial to be FILIAL, got %s", *got.DescricaoMatrizFilial)
	}
	if got.NomeFantasia != "REGIONAL BRASILIA-DF" {
		t.Errorf("expected NomeFantasia to be REGIONAL BRASILIA-DF, got %s", got.NomeFantasia)
	}
	if *got.SituacaoCadastral != 2 {
		t.Errorf("expected SituacaoCadastral to be 2, got %d", *got.SituacaoCadastral)
	}
	if *got.DataSituacaoCadastral != dataSituacaoCadastral {
		t.Errorf("expected SituacaoCadastral to be %v, got %v", dataSituacaoCadastral, *got.DataSituacaoCadastral)
	}
	if *got.DescricaoSituacaoCadastral != "ATIVA" {
		t.Errorf("expected DescricaoSituacaoCadastral to be ATIVA, got %s", *got.DescricaoSituacaoCadastral)
	}
	if *got.MotivoSituacaoCadastral != 0 {
		t.Errorf("expected MotivoSituacaoCadastral to be 0, got %d", *got.MotivoSituacaoCadastral)
	}
	if got.DescricaoMotivoSituacaoCadastral == nil || *got.DescricaoMotivoSituacaoCadastral != "SEM MOTIVO" {
		t.Errorf("expected DescricaoMotivoSituacaoCadastral to be SEM MOTIVO, got %v", got.DescricaoMotivoSituacaoCadastral)
	}
	if got.NomeCidadeNoExterior != "" {
		t.Errorf("expected NomeCidadeNoExterior to be empty, got %s", got.NomeCidadeNoExterior)
	}
	if got.CodigoPais != nil {
		t.Errorf("expected CodigoPais to be nil, got %v", got.CodigoPais)
	}
	if *got.DataInicioAtividade != date(time.Date(1967, 6, 30, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected DataInicioAtividade to be 1967-06-30, got %v", *got.DataInicioAtividade)
	}
	if *got.CNAEFiscal != 6204000 {
		t.Errorf("expected CNAEFiscal to be 6204000, got %d", *got.CNAEFiscal)
	}
	if *got.CNAEFiscalDescricao != "Consultoria em tecnologia da informação" {
		t.Errorf("expected CNAEFiscalDescricao to be Consultoria em tecnologia da informação, got %s", *got.CNAEFiscalDescricao)
	}
	if got.DescricaoTipoDeLogradouro != "AVENIDA" {
		t.Errorf("expected DescricaoTipoDeLogradouro to be AVENIDA, got %s", got.DescricaoTipoDeLogradouro)
	}
	if got.Logradouro != "L2 SGAN" {
		t.Errorf("expected Logradouro to be L2 SGAN, got %s", got.Logradouro)
	}
	if got.Numero != "601" {
		t.Errorf("expected Numero to be 601, got %s", got.Numero)
	}
	if got.Complemento != "MODULO G" {
		t.Errorf("expected Complemento to be MODULO G, got %s", got.Complemento)
	}
	if got.Bairro != "ASA NORTE" {
		t.Errorf("expected Bairro to be ASA NORTE, got %s", got.Bairro)
	}
	if got.CEP != "70836900" {
		t.Errorf("expected CEP to be 70836900, got %s", got.CEP)
	}
	if got.UF != "DF" {
		t.Errorf("expected UF to be DF, got %s", got.UF)
	}
	if *got.CodigoMunicipio != 9701 {
		t.Errorf("expected CodigoMunicipio to be 9701, got %d", *got.CodigoMunicipio)
	}
	if *got.CodigoMunicipioIBGE != 5300108 {
		t.Errorf("expected CodigoMunicipioIBGE to be 5300108, got %d", *got.CodigoMunicipioIBGE)
	}
	if *got.Municipio != "BRASILIA" {
		t.Errorf("expected Municipio to be BRASILIA, got %s", *got.Municipio)
	}
	if got.Telefone1 != "" {
		t.Errorf("expected Telefone1 to be empty, got %s", got.Telefone1)
	}
	if got.Telefone2 != "" {
		t.Errorf("expected Telefone2 to be empty, got %s", got.Telefone2)
	}
	if got.Fax != "" {
		t.Errorf("expected Fax to be empty, got %s", got.Fax)
	}
	if got.Email == nil || *got.Email != "test@ser.pro" {
		t.Errorf("expected Email to be empty string, got %v", got.Email)
	}
	if got.SituacaoEspecial != "" {
		t.Errorf("expected SituacaoEspecial to be empty, got %s", got.SituacaoEspecial)
	}
	if got.DataSituacaoEspecial != nil {
		t.Errorf("expected DataSituacaoEspecial to be nil, got %v", got.DataSituacaoEspecial)
	}
	if len(got.CNAESecundarios) != 5 {
		t.Errorf("expected CNAESecundarios to have 5 items, got %d", len(got.CNAESecundarios))
	}
	if got.RazaoSocial != "SERVICO FEDERAL DE PROCESSAMENTO DE DADOS (SERPRO)" {
		t.Errorf("expected RazaoSocial to be SERVICO FEDERAL DE PROCESSAMENTO DE DADOS (SERPRO), got %s", got.RazaoSocial)
	}
	if *got.CodigoNaturezaJuridica != 2011 {
		t.Errorf("expected CodigoNaturezaJuridica to be 2011, got %d", *got.CodigoNaturezaJuridica)
	}
	if *got.NaturezaJuridica != "Empresa Pública" {
		t.Errorf("expected NaturezaJuridica to be Empresa Pública, got %s", *got.NaturezaJuridica)
	}
	if *got.QualificacaoDoResponsavel != 16 {
		t.Errorf("expected QualificacaoDoResponsavel to be 16, got %d", *got.QualificacaoDoResponsavel)
	}
	if *got.CapitalSocial != 1061004829.23 {
		t.Errorf("expected CapitalSocial to be 1061004829.23, got %f", *got.CapitalSocial)
	}
	if *got.CodigoPorte != 5 {
		t.Errorf("expected CodigoPorte to be 5, got %d", *got.CodigoPorte)
	}
	if *got.Porte != "DEMAIS" {
		t.Errorf("expected Porte to be DEMAIS, got %s", *got.Porte)
	}
	if got.EnteFederativoResponsavel != "" {
		t.Errorf("expected EnteFederativoResponsavel to be empty, got %s", got.EnteFederativoResponsavel)
	}
	if len(got.QuadroSocietario) != 6 {
		t.Errorf("expected QuadroSocietario to have 6 items, got %d items", len(got.QuadroSocietario))
	}
	if len(got.RegimeTributario) != 1 {
		t.Errorf("expected RegimeTributario to have 1 item, got %d items", len(got.RegimeTributario))
	}
	if *got.OpcaoPeloSimples != true {
		t.Errorf("expected OpcaoPeloSimples to be true, got %v", *got.OpcaoPeloSimples)
	}
	if *got.DataOpcaoPeloSimples != date(time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected DataOpcaoPeloSimples to be 2014-01-01, got %v", *got.DataOpcaoPeloSimples)
	}
	if got.DataExclusaoDoSimples != nil {
		t.Errorf("expected DataExclusaoDoSimples to be nil, got %v", got.DataExclusaoDoSimples)
	}
	if *got.OpcaoPeloMEI != false {
		t.Errorf("expected OpcaoPeloMEI to be false, got %v", *got.OpcaoPeloMEI)
	}
	if got.DataOpcaoPeloMEI != nil {
		t.Errorf("expected DataOpcaoPeloMEI to be nil, got %v", got.DataOpcaoPeloMEI)
	}
	if got.DataExclusaoDoMEI != nil {
		t.Errorf("expected DataExclusaoDoMEI to be nil, got %v", got.DataExclusaoDoMEI)
	}
	if got.Pais != nil {
		t.Errorf("expected Pais to be nil, got %v", got.Pais)
	}
	if len(got.QuadroSocietario) != 6 {
		t.Errorf("expected QuadroSocietario to have 6 items, got %d", len(got.QuadroSocietario))
	}
	// Partners are sorted alphabetically by name
	if got.QuadroSocietario[0].NomeSocio != "ANDRE DE CESERO" {
		t.Errorf("expected first partner to be ANDRE DE CESERO, got %s", got.QuadroSocietario[0].NomeSocio)
	}
	if got.QuadroSocietario[0].CNPJCPFDoSocio != "***220050**" {
		t.Errorf("expected first partner CNPJ/CPF to be ***220050**, got %s", got.QuadroSocietario[0].CNPJCPFDoSocio)
	}
	if *got.QuadroSocietario[0].CodigoQualificacaoSocio != 10 {
		t.Errorf("expected partner qualification code to be 10, got %d", *got.QuadroSocietario[0].CodigoQualificacaoSocio)
	}
	if *got.QuadroSocietario[0].QualificaoSocio != "Diretor" {
		t.Errorf("expected partner qualification to be Diretor, got %s", *got.QuadroSocietario[0].QualificaoSocio)
	}
	if *got.QuadroSocietario[0].CodigoFaixaEtaria != 6 {
		t.Errorf("expected partner age range code to be 6, got %d", *got.QuadroSocietario[0].CodigoFaixaEtaria)
	}
	if *got.QuadroSocietario[0].FaixaEtaria != "Entre 51 a 60 anos" {
		t.Errorf("expected partner age range to be Entre 51 a 60 anos, got %s", *got.QuadroSocietario[0].FaixaEtaria)
	}
	if len(got.RegimeTributario) != 1 {
		t.Errorf("expected RegimeTributario to have 1 item, got %d", len(got.RegimeTributario))
	}
	if got.RegimeTributario[0].Ano != 2018 {
		t.Errorf("expected tax regime year to be 2018, got %d", got.RegimeTributario[0].Ano)
	}
	if got.RegimeTributario[0].FormaDeTributação != "LUCRO PRESUMIDO" {
		t.Errorf("expected tax regime type to be LUCRO PRESUMIDO, got %s", got.RegimeTributario[0].FormaDeTributação)
	}
	if got.RegimeTributario[0].QuantidadeDeEscrituracoes != 1 {
		t.Errorf("expected tax regime quantity to be 1, got %d", got.RegimeTributario[0].QuantidadeDeEscrituracoes)
	}
}

func TestNewCompanyWithPrivacy(t *testing.T) {
	kv, err := newBadger(t.TempDir(), false)
	if err != nil {
		t.Fatalf("expected no error creating kv, got %s", err)
	}
	srcs := sources()
	ctx := context.Background()
	for key, src := range srcs {
		if key == "est" {
			continue
		}
		if err := loadCSVs(ctx, "../testdata", src, nil, kv); err != nil {
			t.Fatalf("expected no error loading %s data, got %s", key, err)
		}
	}
	row := []string{
		"33683111",               // 0 CNPJ Base
		"0002",                   // 1 CNPJ Ordem
		"80",                     // 2 CNPJ DV
		"1",                      // 3 Indentificador Matriz/Filial (MATRIZ)
		"João Silva 12345678901", // 4 Nome Fantasia with CPF
		"02",                     // 5 Situação Cadastral
		"20040522",               // 6 Data Situação Cadastral
		"00",                     // 7 Motivo Situação Cadastral
		"",                       // 8 Nome da cidade no exterior
		"",                       // 9 Pais
		"19670630",               // 10 Data de Início da Ativiade
		"6204000",                // 11 CNAE Fiscal
		"",                       // 12 CNAEs Secundários
		"RUA",                    // 13 Tipo de Logradouro
		"L2 SGAN",                // 14 Logradouro
		"601",                    // 15 Número
		"MODULO G",               // 16 Complemento
		"ASA NORTE",              // 17 Bairro
		"70836900",               // 18 CEP
		"DF",                     // 19 UF
		"9701",                   // 20 Município
		"61",                     // 21 DDD 1
		"12345678",               // 22 Telefone 1
		"",                       // 23 DDD 2
		"87654321",               // 24 Telefone 2
		"11",                     // 25 DDD Fax
		"",                       // 26 Fax
		"test@example.com",       // 27 Email
		"",                       // 28 Situação Especial
		"",                       // 29 Data Situação Especial
	}
	got, err := newCompany(srcs, kv, row)
	if err != nil {
		t.Fatalf("expected no error creating a company, got %s", err)
	}
	got.withPrivacy()
	if got.Email != nil {
		t.Errorf("expected Email to be nil after privacy, got %v", got.Email)
	}
	// SERPRO is a public company (Empresa Pública), not an individual
	// So address fields should NOT be cleared
	if got.DescricaoTipoDeLogradouro != "RUA" {
		t.Errorf("expected DescricaoTipoDeLogradouro to be RUA for public company, got %s", got.DescricaoTipoDeLogradouro)
	}
	if got.Logradouro != "L2 SGAN" {
		t.Errorf("expected Logradouro to be L2 SGAN for public company, got %s", got.Logradouro)
	}
	if got.Numero != "601" {
		t.Errorf("expected Numero to be 601 for public company, got %s", got.Numero)
	}
	if got.Complemento != "MODULO G" {
		t.Errorf("expected Complemento to be MODULO G for public company, got %s", got.Complemento)
	}
	if got.Telefone1 != "6112345678" {
		t.Errorf("expected Telefone1 to be 6112345678 for public company, got %s", got.Telefone1)
	}
	if got.Telefone2 != "87654321" {
		t.Errorf("expected Telefone2 to be 87654321 for public company, got %s", got.Telefone2)
	}
	if got.Fax != "11" {
		t.Errorf("expected Fax to be 11 for public company, got %s", got.Fax)
	}
	want := "João Silva ***45678***"
	if got.NomeFantasia != want {
		t.Errorf("expected NomeFantasia to be %s after privacy, got %s", want, got.NomeFantasia)
	}
}
