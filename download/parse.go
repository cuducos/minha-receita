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
		l[3:17],    // CNPJ
		l[17:18],   // IdentificadorDeSocio
		l[18:168],  // NomeSocio
		l[168:182], // CNPJCPFDoSocio
		l[182:184], // CodigoQualificacaoSocio
		l[184:189], // PercentualCapitalSocial
		l[189:197], // DataEntradaSociedade
		l[270:281], // CPFRepresentanteLegal
		l[281:341], // NomeRepresentanteLegal
		l[341:343], // CodigoQualificacaoRepresentanteLegal
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
		l[3:17],    //  0 CNPJ
		l[17:18],   //  1 IdentificadorMatrizFilial
		l[18:168],  //  2 RazaoSocial
		l[168:223], //  3 NomeFantasia
		l[223:225], //  4 SituacaoCadastral
		l[225:233], //  5 DataSituacaoCadastral
		l[233:235], //  6 MotivoSituacaoCadastral
		l[235:290], //  7 NomeCidadeExterior
		l[362:367], //  8 CodigoNaturezaJuridica
		l[367:375], //  9 DataInicioAtividade
		l[375:382], // 10 CNAEFiscal
		l[382:402], // 11 DescricaoTipoLogradouro
		l[402:462], // 12 Logradouro
		l[462:467], // 13 Numero
		l[467:624], // 14 Complemento
		l[624:674], // 15 Bairro
		l[674:682], // 16 CEP
		l[682:684], // 17 UF
		l[684:688], // 18 CodigoMunicipio
		l[688:738], // 29 Municipio
		l[738:750], // 20 DDDTelefone1
		l[750:762], // 21 DDDTelefone2
		l[762:774], // 22 DDDFax
		l[889:891], // 23 QualificacaoDoResponsavel
		l[891:905], // 24 CapitalSocial
		l[905:907], // 25 Porte
		l[907:908], // 26 OpcaoPeloSimples
		l[908:916], // 27 DataOpcaoPeloSimples
		l[916:924], // 28 DataExclusaoDoSimples
		l[924:925], // 29 OpcaoPeloMEI
		l[925:948], // 30 SituacaoEspecial
		l[948:956], // 31 DataSituacaoEspecial
	})

	// format dates
	for _, i := range []int{5, 9, 27, 28, 31} {
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
