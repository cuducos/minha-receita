package download

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFederalRevenueGetURLs(t *testing.T) {
	tmp := t.TempDir()
	ts := httpTestServer(t, "cadastro-nacional-de-pessoa-juridica-cnpj.json")
	defer ts.Close()

	t.Run("returns download urls", func(t *testing.T) {
		got, err := federalRevenueGetURLs(ts.URL, tmp)
		if err != nil {
			t.Errorf("expected to run without errors, got: %v:", err)
		}
		expected := []string{
			"https://dadosabertos.rfb.gov.br/CNPJ/Cnaes.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas0.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas1.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas2.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas3.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas4.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas5.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas6.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas7.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas8.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Empresas9.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos0.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos1.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos2.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos3.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos4.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos5.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos6.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos7.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos8.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Estabelecimentos9.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Motivos.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Municipios.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Naturezas.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Paises.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Qualificacoes.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Simples.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios0.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios1.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios2.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios3.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios4.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios5.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios6.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios7.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios8.zip",
			"https://dadosabertos.rfb.gov.br/CNPJ/Socios9.zip",
		}
		assertArraysHaveSameItems(t, got, expected)
	})

	t.Run("saves updated at date", func(t *testing.T) {
		_, err := federalRevenueGetURLs(ts.URL, tmp)
		if err != nil {
			t.Errorf("expected to run without errors, got: %v:", err)
		}
		pth := filepath.Join(tmp, FederalRevenueUpdatedAt)
		got, err := os.ReadFile(pth)
		if err != nil {
			t.Errorf("expected no error reading %s, updatedAt %s", pth, err)
		}
		expected := "2024-04-13"
		if string(got) != expected {
			t.Errorf("expected updated at to be %s, got %s", expected, string(got))
		}
	})
}
