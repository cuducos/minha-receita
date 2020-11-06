// Package download privacy decisions are heavily based on Álvaro's “Turicas”
// Justen's “Sócios de Empresas Brasileiras” decisions and design:
// https://github.com/turicas/socios-brasil — licensed under LGPL-3.0 License.
package download

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

const cleanCompanyNamePattern = `( ?-? ?(CPF)? ?(\d{11})?)$`

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func allZeros(s string) bool {
	v, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return 0 == v
}

func removeLeftZeros(s string) string {
	return strings.TrimLeft(strings.TrimSpace(s), "0")
}

func formatDate(s string) string {
	d, err := time.Parse("20060102", s)
	if err != nil {
		return s
	}
	return d.Format("2006-01-02")
}

func cleanLine(l []string) []string {
	var c []string
	for _, v := range l {
		c = append(c, strings.TrimSpace(v))
	}
	return c
}

func cleanCompanyName(n string) string {
	if n == "" || isNumeric(n) {
		return n
	}
	r := regexp.MustCompile(cleanCompanyNamePattern)
	return strings.TrimSpace(r.ReplaceAllString(n, ""))
}

func parsePartner(l string) []string {
	cols := cleanLine([]string{
		l[3:17],    // cnpj
		l[17:18],   // identificadorDeSocio
		l[18:168],  // nomeSocio
		l[168:182], // cnpjCpfDoSocio
		l[182:184], // codigoQualificacaoSocio
		l[184:189], // percentualCapitalSocial
		l[189:197], // dataEntradaSociedade
		l[270:281], // cpfRepresentanteLegal
		l[281:341], // nomeRepresentanteLegal
		l[341:343], // codigoQualificacaoRepresentanteLegal
	})

	// fix date
	cols[6] = formatDate(cols[6])

	// remove placeholder in case of CPF
	if strings.HasPrefix(cols[3], "000") {
		cols[3] = strings.TrimPrefix(cols[3], "000")
	}

	// delete unnused data
	cols[5] = removeLeftZeros(cols[5])
	cols[9] = removeLeftZeros(cols[9])
	if cols[8] == "CPF INVALIDO" {
		cols[7] = ""
		cols[8] = ""
	}

	return cols
}

func parseCNAE(l string) [][]string {
	var sc [][]string
	cnpj := l[3:17]

	for i := 17; i < 711; i = i + 7 {
		c := strings.TrimSpace(l[i : i+7])
		if c != "" && !allZeros(c) {
			c = removeLeftZeros(c)
			sc = append(sc, []string{cnpj, c})
		}
	}
	return sc
}

func parseCompany(l string) []string {
	cols := cleanLine([]string{
		l[3:17],    // Cnpj
		l[17:18],   // IdentificadorMatrizFilial
		l[18:168],  // RazaoSocial
		l[168:223], // NomeFantasia
		l[223:225], // SituacaoCadastral
		l[225:233], // DataSituacaoCadastral
		l[233:235], // MotivoSituacaoCadastral
		l[235:290], // NomeCidadeExterior
		l[362:367], // CodigoNaturezaJuridica
		l[367:375], // DataInicioAtividade
		l[375:382], // CNAEFiscal
		l[382:402], // DescricaoTipoLogradouro
		l[402:462], // Logradouro
		l[462:467], // Numero
		l[467:624], // Complemento
		l[624:674], // Bairro
		l[674:682], // Cep
		l[682:684], // Uf
		l[684:688], // CodigoMunicipio
		l[688:738], // Municipio
		l[738:750], // DddTelefone1
		l[750:762], // DddTelefone2
		l[762:774], // DddFax
		l[889:891], // QualificacaoDoResponsavel
		l[891:905], // CapitalSocial
		l[905:907], // Porte
		l[907:908], // OpcaoPeloSimples
		l[908:916], // DataOpcaoPeloSimples
		l[916:924], // DataExclusaoDoSimples
		l[924:925], // OpcaoPeloMei
		l[925:948], // SituacaoEspecial
		l[948:956], // DataSituacaoEspecial
	})

	// format dates
	for _, i := range []int{5, 9} {
		cols[i] = formatDate(cols[i])
	}

	// format booleans
	if cols[26] == "5" || cols[26] == "7" {
		cols[26] = "true"
	} else {
		cols[26] = "false"
	}
	if cols[29] == "S" {
		cols[29] = "true"
	} else {
		cols[29] = "false"
	}

	// delete unnused data
	for _, i := range []int{4, 6, 24, 25, 27, 28, 31} {
		cols[i] = removeLeftZeros(cols[i])
	}

	// privacy issues
	deleteFromIndividuals := []int{
		11, //	descricao_tipo_logradouro
		12, //	logradouro
		13, //	numero
		14, //	complemento
		20, //	ddd_telefone_1
		21, //	ddd_telefone_2
		22, //	ddd_fax
	}
	individualCompanies := []string{"2135", "2305", "2305", "4014", "4081"}
	isIndividual := false
	for _, i := range individualCompanies {
		if i == cols[8] {
			isIndividual = true
		}
	}
	if isIndividual {
		for _, i := range deleteFromIndividuals {
			cols[i] = ""
		}
		cols[2] = cleanCompanyName(cols[2])
		cols[3] = cleanCompanyName(cols[3])
	}

	return cols
}

type parsedLine struct {
	valid    bool
	kind     string
	contents [][]string
}

func parseLine(l string) parsedLine {
	var p parsedLine
	switch l[0:1] {
	case "1":
		p.valid = true
		p.kind = "empresa"
		p.contents = [][]string{parseCompany(l)}
	case "2":
		p.valid = true
		p.kind = "socio"
		p.contents = [][]string{parsePartner(l)}
	case "6":
		p.valid = true
		p.kind = "cnae"
		p.contents = parseCNAE(l)
	}
	return p
}

func parseZipFile(wg *sync.WaitGroup, ch chan<- parsedLine, z *zippedFile) {
	wg.Add(1)
	defer wg.Done()
	defer z.closeReaders()

	s := bufio.NewScanner(z.firstFile)
	for s.Scan() {
		l := parseLine(s.Text())
		if l.valid {
			ch <- l
		}
	}
}

func status(w *writers, f, c int) error {
	company, err := os.Stat(w.company.path)
	if err != nil {
		return err
	}
	partner, err := os.Stat(w.partner.path)
	if err != nil {
		return err
	}
	cnae, err := os.Stat(w.cnae.path)
	if err != nil {
		return err
	}

	fmt.Printf(
		"\rFixed-width lines read: %s | CSV lines written: %s | %s: %s | %s: %s | %s: %s ",
		humanize.Comma(int64(f)),
		humanize.Comma(int64(c)),
		w.company.path,
		humanize.Bytes(uint64(company.Size())),
		w.partner.path,
		humanize.Bytes(uint64(partner.Size())),
		w.cnae.path,
		humanize.Bytes(uint64(cnae.Size())),
	)
	return nil
}

// Parse the downloaded files and saves a compressed CSV version of them.
func Parse(dir string) {
	w, err := newWriters(dir)
	if err != nil {
		panic(err)
	}
	defer w.closeResources()

	var wg sync.WaitGroup
	lines := make(chan parsedLine)
	for i := 1; i >= 1; i++ { // infinite loop: breaks when file does not exist
		z, err := newZippedFile(dir, i)
		if os.IsNotExist(err) {
			break // no more files to read
		}
		if err != nil {
			panic(err)
		}

		go parseZipFile(&wg, lines, z)
	}

	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(lines)
	}(&wg)

	// show the status (progress)
	var f, c int
	go func() {
		for {
			status(w, f, c)
			time.Sleep(3 * time.Second)
		}
	}()

	for l := range lines {
		switch l.kind {
		case "empresa":
			w.company.write(l.contents)
		case "socio":
			w.partner.write(l.contents)
		case "cnae":
			w.cnae.write(l.contents)
		}
		f++
		c += len(l.contents)
	}
}
