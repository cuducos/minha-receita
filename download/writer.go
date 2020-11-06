package download

import (
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const maxCSVBufferSize = 10_000

type resourceWriter struct {
	header []string
	path   string
	file   *os.File
	gz     *gzip.Writer
	csv    *csv.Writer
	buffer int
}

func (r *resourceWriter) write(ls [][]string) (int, error) {
	if r.csv == nil {
		return 0, errors.New("cannot write to a non-initialized CSV writer")
	}
	for i, l := range ls {
		if len(l) != len(r.header) {
			return 0, fmt.Errorf(
				"cannot write to CSV; the CSV has %d columns, but the line %d has %d columns",
				len(r.header),
				i+1,
				len(l),
			)
		}
	}

	var c int
	for _, l := range ls {
		err := r.csv.Write(l)
		if err != nil {
			return 0, err
		}
		c++
		r.buffer++
	}
	if r.buffer >= maxCSVBufferSize {
		r.csv.Flush()
		if err := r.csv.Error(); err != nil {
			return 0, err
		}
		r.buffer = 0
	}
	return len(ls), nil
}

func (r *resourceWriter) closeFiles() error {
	r.csv.Flush()
	r.gz.Close()
	r.file.Close()
	return r.csv.Error()
}

func newResourceWriter(p string, h []string) (*resourceWriter, error) {
	f, err := os.Create(p)
	if err != nil {
		return nil, err
	}

	r := resourceWriter{
		header: h,
		path:   p,
		file:   f,
		gz:     gzip.NewWriter(f),
	}
	r.csv = csv.NewWriter(r.gz)
	_, err = r.write([][]string{h})
	if err != nil {
		return nil, err
	}
	return &r, nil
}

type writers struct {
	company *resourceWriter
	partner *resourceWriter
	cnae    *resourceWriter
}

func (w *writers) all() []*resourceWriter {
	return []*resourceWriter{w.company, w.partner, w.cnae}
}

func (w *writers) closeResources() {
	for _, r := range w.all() {
		r.closeFiles()
	}
}

func newWriters(dir string) (*writers, error) {
	company, err := newResourceWriter(
		filepath.Join(dir, "empresa.csv.gz"),
		[]string{
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
	)
	if err != nil {
		return nil, err
	}

	partner, err := newResourceWriter(
		filepath.Join(dir, "socio.csv.gz"),
		[]string{
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
	)
	if err != nil {
		return nil, err
	}

	cnae, err := newResourceWriter(
		filepath.Join(dir, "cnae_secundaria.csv.gz"),
		[]string{"cnpj", "cnae"},
	)
	if err != nil {
		return nil, err
	}

	return &writers{company, partner, cnae}, nil
}
