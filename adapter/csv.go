package adapter

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

const separator = ';'

func cleanLine(l []string) ([]string, error) {
	var c []string
	var err error

	for _, v := range l {
		v := strings.TrimSpace(v)
		if !utf8.ValidString(v) {
			v, err = charmap.ISO8859_1.NewDecoder().String(v)
			if err != nil {
				return nil, err
			}
		}
		c = append(c, strings.TrimSpace(v))
	}
	return c, nil
}

func (a *Adapter) lineProducer(l chan<- []string, f *zip.File) error {
	z, err := f.Open()
	if err != nil {
		return err
	}
	defer z.Close()

	r := csv.NewReader(z)
	r.Comma = separator
	for {
		s, err := r.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		l <- s
	}
}

func (a *Adapter) lineConsumer(l chan []string) {
	for {
		s, ok := <-l
		if !ok {
			return
		}

		if err := a.csvWriter.Write(s); err != nil {
			log.Fatal(err)
		}
	}
}

func headersFor(a *Adapter) ([]string, error) {
	var h []string
	switch a.kind {
	case city, cnae, motive, nature, country, qualification:
		h = []string{"codigo", "descricao"}
	case company:
		h = []string{
			"cnpj",
			"razao_social",
			"natureza_juridica",
			"qualificacao_do_responsavel",
			"capital_social",
			"porte",
			"ente_federativo_responsavel",
		}
	case facility:
		h = []string{
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
		h = []string{
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
	case simple:
		h = []string{
			"cnpj",
			"opcao_pelo_simples",
			"data_opcao_pelo_simples",
			"data_exclusao_do_simples",
			"opcao_pelo_mei",
			"data_opcao_pelo_mei",
			"data_entrada_do_mei",
		}
	default:
		return h, fmt.Errorf("No headers found for %s", a.kind)
	}

	return h, nil
}

func (a *Adapter) csvPath() string {
	var n string
	switch a.kind {

	case city:
		n = "municipio.csv"
	case cnae:
		n = "cnae.csv"
	case company:
		n = "empresa.csv"
	case country:
		n = "pais.csv"
	case facility:
		n = "estabelecimento.csv"
	case motive:
		n = "motivo_situacao_cadastral.csv"
	case nature:
		n = "natureza_juridica.csv"
	case partner:
		n = "socio.csv"
	case qualification:
		n = "qualificacao_de_socio.csv"
	case simple:
		n = "simples.csv"
	}

	if a.compression != "" {
		n += "." + a.compression
	}

	return filepath.Join(a.dir, n)
}

func (a *Adapter) createCsv() error {
	p := a.csvPath()
	if err := os.RemoveAll(p); err != nil {
		return err
	}

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	a.fileHandler = f

	w, err := a.Writer(f)
	if err != nil {
		return err
	}
	a.ioWriter = w

	h, err := headersFor(a)
	if err != nil {
		return err
	}

	a.csvWriter = csv.NewWriter(w)
	if err := a.csvWriter.Write(h); err != nil {
		return err
	}

	return nil
}

func (a *Adapter) writeCsv(q chan<- error) {
	ls, err := a.files()
	if err != nil {
		q <- err
		return
	}

	if err := a.createCsv(); err != nil {
		q <- err
		return
	}
	defer a.Close()

	e := make(chan error)
	l := make(chan []string)
	for _, f := range ls {
		go a.unzip(e, l, f)
	}

	go a.lineConsumer(l)
	for range ls {
		err := <-e
		if err != nil {
			q <- err
			return
		}
	}
	close(l)

	q <- nil
	return
}
