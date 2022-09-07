package download

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFederalRevenueGetURLs(t *testing.T) {
	tmp := t.TempDir()
	ts := httpTestServer(t, "dados-publicos-cnpj.html")
	defer ts.Close()

	t.Run("returns download urls", func(t *testing.T) {
		got, err := federalRevenueGetURLs(ts.Client(), ts.URL, tmp)
		if err != nil {
			t.Errorf("expected to run withour errors, got: %v:", err)
		}
		expected := []string{
			"http://200.152.38.155/CNPJ/Cnaes.zip",
			"http://200.152.38.155/CNPJ/Empresas0.zip",
			"http://200.152.38.155/CNPJ/Empresas1.zip",
			"http://200.152.38.155/CNPJ/Empresas2.zip",
			"http://200.152.38.155/CNPJ/Empresas3.zip",
			"http://200.152.38.155/CNPJ/Empresas4.zip",
			"http://200.152.38.155/CNPJ/Empresas5.zip",
			"http://200.152.38.155/CNPJ/Empresas6.zip",
			"http://200.152.38.155/CNPJ/Empresas7.zip",
			"http://200.152.38.155/CNPJ/Empresas8.zip",
			"http://200.152.38.155/CNPJ/Empresas9.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos0.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos1.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos2.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos3.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos4.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos5.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos6.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos7.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos8.zip",
			"http://200.152.38.155/CNPJ/Estabelecimentos9.zip",
			"http://200.152.38.155/CNPJ/Motivos.zip",
			"http://200.152.38.155/CNPJ/Municipios.zip",
			"http://200.152.38.155/CNPJ/Naturezas.zip",
			"http://200.152.38.155/CNPJ/Paises.zip",
			"http://200.152.38.155/CNPJ/Qualificacoes.zip",
			"http://200.152.38.155/CNPJ/Simples.zip",
			"http://200.152.38.155/CNPJ/Socios0.zip",
			"http://200.152.38.155/CNPJ/Socios1.zip",
			"http://200.152.38.155/CNPJ/Socios2.zip",
			"http://200.152.38.155/CNPJ/Socios3.zip",
			"http://200.152.38.155/CNPJ/Socios4.zip",
			"http://200.152.38.155/CNPJ/Socios5.zip",
			"http://200.152.38.155/CNPJ/Socios6.zip",
			"http://200.152.38.155/CNPJ/Socios7.zip",
			"http://200.152.38.155/CNPJ/Socios8.zip",
			"http://200.152.38.155/CNPJ/Socios9.zip",
			"http://200.152.38.155/CNPJ/anual/Dados%20Abertos%20S%c3%adtio%20RFB%20Extracao%2020.10.2021.zip",
		}
		assertArraysHaveSameItems(t, got, expected)
	})

	t.Run("saves updated at date", func(t *testing.T) {
		_, err := federalRevenueGetURLs(ts.Client(), ts.URL, tmp)
		if err != nil {
			t.Errorf("expected to run withour errors, got: %v:", err)
		}
		pth := filepath.Join(tmp, FederalRevenueUpdatedAt)
		got, err := os.ReadFile(pth)
		if err != nil {
			t.Errorf("expected no error reading %s, updatedAt %s", pth, err)
		}
		expected := "2022-07-09"
		if string(got) != expected {
			t.Errorf("expected updated at to be %s, got %s", expected, string(got))
		}
	})
}
