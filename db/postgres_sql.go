package db

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/go-pg/pg/v10"
)

type source struct {
	name   string
	schema string
}

func (s source) path(dir string) string {
	var f string
	if s.name == "cnae" {
		f = "CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx"
	} else {
		f = fmt.Sprintf("%s.csv.gz", s.name)
	}

	p, err := filepath.Abs(filepath.Clean(filepath.Join(dir, f)))
	if err != nil {
		panic(err)
	}
	return p
}

func (s source) tableName() string {
	return fmt.Sprintf("%ss", s.name)
}

func (s source) fullTableName() string {
	return fmt.Sprintf("%s.%s", s.schema, s.tableName())
}

func (s source) indexName() string {
	if s.name == "cnae" {
		return "codigo"
	}
	return "cnpj"
}

func (s source) columns() string {
	var cols []string
	switch s.name {
	case "cnae":
		cols = []string{"codigo int8 NOT NULL", "descricao text NOT NULL"}
	case "cnae_secundaria":
		cols = []string{"cnpj text NOT NULL", "cnae int8 NOT NULL"}
	case "socio":
		cols = []string{
			"cnpj text NOT NULL",
			"identificador_de_socio int8 NULL",
			"nome_socio text NULL",
			"cnpj_cpf_do_socio text NULL",
			"codigo_qualificacao_socio int8 NULL",
			"percentual_capital_social int8 NULL",
			"data_entrada_sociedade date NULL",
			"cpf_representante_legal text NULL",
			"nome_representante_legal text NULL",
			"codigo_qualificacao_representante_legal int8 NULL",
		}
	case "empresa":
		cols = []string{
			"cnpj text NOT NULL",
			"identificador_matriz_filial integer NULL",
			"razao_social text NULL",
			"nome_fantasia text NULL",
			"situacao_cadastral integer NULL",
			"data_situacao_cadastral date NULL",
			"motivo_situacao_cadastral integer NULL",
			"nome_cidade_exterior text NULL",
			"codigo_natureza_juridica integer NULL",
			"data_inicio_atividade date NULL",
			"cnae_fiscal integer NULL",
			"descricao_tipo_logradouro text NULL",
			"logradouro text NULL",
			"numero text NULL",
			"complemento text NULL",
			"bairro text NULL",
			"cep text NULL",
			"uf text NULL",
			"codigo_municipio integer NULL",
			"municipio text NULL",
			"ddd_telefone1 text NULL",
			"ddd_telefone2 text NULL",
			"ddd_fax text NULL",
			"qualificacao_do_responsavel integer NULL",
			"capital_social decimal NULL",
			"porte integer NULL",
			"opcao_pelo_simples boolean NOT NULL",
			"data_opcao_pelo_simples text NULL",
			"data_exclusao_do_simples text NULL",
			"opcao_pelo_mei boolean NOT NULL",
			"situacao_especial text NULL",
			"data_situacao_especial text NULL",
		}
	default:
		panic(fmt.Sprintf("No columns defined for source %s", s.name))
	}
	return strings.Join(cols, ", ")
}

func getSources(schema string) []source {
	var s []source
	for _, n := range []string{"empresa", "cnae", "cnae_secundaria", "socio"} {
		s = append(s, source{n, schema})
	}
	return s
}

func createTable(db *pg.DB, wg *sync.WaitGroup, s source) {
	defer wg.Done()
	t := s.fullTableName()
	i := s.indexName()
	log.Output(2, fmt.Sprintf("Creating table %s…", t))
	_, err := db.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (%s);
		CREATE INDEX IF NOT EXISTS idx_%s_%s ON %s USING btree (%s);
	`, t, s.columns(), s.tableName(), i, t, i))
	if err != nil {
		panic(err)
	}
}

func dropTable(db *pg.DB, wg *sync.WaitGroup, s source) {
	defer wg.Done()
	t := s.fullTableName()
	log.Output(2, fmt.Sprintf("Dropping table %s…", t))
	_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", t))
	if err != nil {
		panic(err)
	}
}

func copyFrom(db *pg.DB, wg *sync.WaitGroup, s source, dir string) {
	defer wg.Done()
	table := s.fullTableName()
	src := s.path(dir)
	log.Output(2, fmt.Sprintf("Importing data from %s to %s…", src, table))
	defer log.Output(2, fmt.Sprintf("Done! Imported data from %s to %s.", src, table))

	f, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	cmd := exec.Command(
		"psql",
		pgURI,
		"-c",
		fmt.Sprintf(`\copy %s FROM STDIN DELIMITER ',' CSV HEADER;`, table),
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	defer stdin.Close()

	cmd.Start()
	_, err = io.Copy(stdin, r)
	if err != nil {
		panic(err)
	}
}

func parseCNAE(r []string) (CNAE, error) {
	var c CNAE
	var err error

	r[4] = regexp.MustCompile(`\D`).ReplaceAllString(r[4], "") // remove non-digit
	if r[4] == "" {
		return c, errors.New("Código inválido")
	}

	c.Codigo, err = strconv.Atoi(r[4])
	if err != nil {
		return c, errors.New("Código inválido")
	}

	c.Descricao = r[5]
	return c, nil
}

func importCNAEXls(db *pg.DB, wg *sync.WaitGroup, s source, dir string) {
	defer wg.Done()
	p := s.path(dir)
	t := s.fullTableName()
	log.Output(2, fmt.Sprintf("Importing data from %s to %s…", p, t))
	defer log.Output(2, fmt.Sprintf("Done! Imported data from %s to %s.", p, t))

	f, err := excelize.OpenFile(p)
	if err != nil {
		panic(err)
	}

	rows, err := f.GetRows("Estrutura Det. CNAE Subclass2.3")
	if err != nil {
		panic(err)
	}

	sql := fmt.Sprintf("INSERT INTO %s VALUES ", t)
	for _, r := range rows {
		row, err := parseCNAE(r)
		if err != nil { // just skip the line
			continue
		}
		sql += fmt.Sprintf("(%d, '%s'),", row.Codigo, row.Descricao)
	}
	sql = strings.TrimSuffix(sql, ",")
	sql += ";"

	_, err = db.Exec(sql)
	if err != nil {
		panic(err)
	}
}

func queryPartners(db *pg.DB, wg *sync.WaitGroup, c *Company) {
	defer wg.Done()
	_, err := db.Query(&c.QSA, "SELECT * FROM socios WHERE cnpj = ?", c.CNPJ)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get partners for %s: %v", c, err)
	}
}

func queryActivities(db *pg.DB, wg *sync.WaitGroup, c *Company) {
	defer wg.Done()
	_, err := db.Query(&c.CNAESecundarias, `
		SELECT cnae_secundarias.cnae AS codigo, cnaes.descricao
		FROM cnae_secundarias
		INNER JOIN cnaes ON cnae_secundarias.cnae = cnaes.codigo
		WHERE cnpj = ?
	`, c.CNPJ)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get secondary CNAE for %s: %v", c, err)
	}
}

func getCompany(db *pg.DB, num string) (Company, error) {
	var c Company
	_, err := db.QueryOne(&c, `
        SELECT empresas.*, cnaes.descricao AS cnae_fiscal_descricao
        FROM empresas
        LEFT JOIN cnaes ON empresas.cnae_fiscal = cnaes.codigo
        WHERE cnpj = ?
	`, num)
	if err != nil {
		return c, err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go queryPartners(db, &wg, &c)
	go queryActivities(db, &wg, &c)
	wg.Wait()
	return c, nil
}
