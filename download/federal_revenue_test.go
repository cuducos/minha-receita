package download

import "testing"

func TestFederalRevenueGetMostRecentURL(t *testing.T) {
	ts := httpTestServer(t, []string{"dados_abertos_cnpj.html"})
	defer ts.Close()

	t.Run("returns download urls", func(t *testing.T) {
		got, err := federalRevenueGetMostRecentURL(ts.URL)
		if err != nil {
			t.Errorf("expected to run without errors, got: %v:", err)
		}
		expected := ts.URL + "/2024-08/"
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})
}

func TestFederalRevenueGetURLs(t *testing.T) {
	ts := httpTestServer(t, []string{"dados_abertos_cnpj.html", "2024-08.html", "regime_tributario.html"})
	defer ts.Close()

	t.Run("returns download urls", func(t *testing.T) {
		got, err := federalRevenueGetURLs(ts.URL)
		if err != nil {
			t.Errorf("expected to run without errors, got: %v:", err)
		}
		expected := []string{
			ts.URL + "/dados_abertos_cnpj/2024-08/Cnaes.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas0.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas1.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas2.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas3.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas4.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas5.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas6.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas7.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas8.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Empresas9.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos0.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos1.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos2.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos3.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos4.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos5.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos6.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos7.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos8.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Estabelecimentos9.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Motivos.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Municipios.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Naturezas.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Paises.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Qualificacoes.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Simples.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios0.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios1.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios2.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios3.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios4.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios5.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios6.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios7.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios8.zip",
			ts.URL + "/dados_abertos_cnpj/2024-08/Socios9.zip",
			ts.URL + "/regime_tributario/Lucro%20Arbitrado.zip",
			ts.URL + "/regime_tributario/Lucro%20Presumido.zip",
			ts.URL + "/regime_tributario/Lucro%20Real.zip",
			ts.URL + "/regime_tributario/Imunes%20e%20Isentas.zip",
		}
		assertArraysHaveSameItems(t, got, expected)
	})
}
