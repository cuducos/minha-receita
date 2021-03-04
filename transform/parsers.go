package transform

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/frictionlessdata/tableschema-go/schema"
	"golang.org/x/text/encoding/charmap"
)

const cleanCompanyNamePattern = `( ?-? ?(CPF)? ?(\d{11})?)$`

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// values comes as integers (to preserve precision, I guess) so we need to
// format them as decimals (e.g.: "123456" => "1234.56")
func addDecimalSeparator(s string) string {
	if s == "" || s == "0" {
		return s
	}
	if len(s) <= 2 {
		return "0." + s
	}

	i := s[0 : len(s)-2]
	d := s[len(s)-2:]
	return fmt.Sprintf("%s.%s", i, d)
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
	if allZeros(s) {
		return ""
	}

	d, err := time.Parse("20060102", s)
	if err != nil {
		return ""
	}
	return d.Format("2006-01-02")
}

func cleanLine(l []string) []string {
	var c []string
	var err error

	for _, v := range l {
		v := strings.TrimSpace(v)
		if !utf8.ValidString(v) {
			v, err = charmap.ISO8859_1.NewDecoder().String(v)
			if err != nil {
				panic(err)
			}
		}
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
		l[3:17],    // 0 CNPJ
		l[17:18],   // 1 IdentificadorDeSocio
		l[18:168],  // 2 NomeSocio
		l[168:182], // 3 CNPJCPFDoSocio
		l[182:184], // 4 CodigoQualificacaoSocio
		l[184:189], // 5 PercentualCapitalSocial
		l[189:197], // 6 DataEntradaSociedade
		l[270:281], // 7 CPFRepresentanteLegal
		l[281:341], // 8 NomeRepresentanteLegal
		l[341:343], // 9 CodigoQualificacaoRepresentanteLegal
	})

	// format date
	cols[6] = formatDate(cols[6])

	// delete unnused data
	cols[3] = removeLeftZeros(cols[3])
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

	// delete unnused data
	for _, i := range []int{4, 6, 24, 25, 27, 28, 31} {
		cols[i] = removeLeftZeros(cols[i])
	}

	// properly format decimals
	cols[24] = addDecimalSeparator(cols[24])

	// privacy issues
	deleteFromIndividuals := []int{
		11, // descricao_tipo_logradouro
		12, // logradouro
		13, // numero
		14, // complemento
		20, // ddd_telefone_1
		21, // ddd_telefone_2
		22, // ddd_fax
	}
	individualCompanies := []string{"2135", "2305", "2305", "4014", "4081"}
	isIndividual := false
	for _, i := range individualCompanies {
		if i == cols[8] {
			isIndividual = true
			break
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

func parseField(v string, f schema.Field) string {
	// return empty string if value is in missing values slice
	if _, ok := f.MissingValues[v]; ok {
		return ""
	}

	// serialize boolean field
	for _, b := range f.TrueValues {
		if v == b {
			return "true"
		}
	}
	for _, b := range f.FalseValues {
		if v == b {
			return "false"
		}
	}

	// remove null character
	return strings.Replace(v, "\x00", "", -1)
}

type parsedLine struct {
	src      string
	valid    bool
	kind     string
	contents [][]string
}

func (p *parsedLine) validate(s Schema) {
	p.valid = true

	// check whether we have contents
	if len(p.contents) == 0 {
		p.valid = false
		return
	}

	// check if contents have the same number of columns
	for _, l := range p.contents {
		if len(l) != len(s.Fields) {
			p.valid = false
		}
	}
	if !p.valid {
		return
	}

	// check whether we can cast each column to the schema data type
	for _, l := range p.contents {
		for i, f := range s.Fields {
			if l[i] == "" && !f.Constraints.Required {
				continue
			}
			_, err := f.Cast(l[i])
			if err != nil {
				log.Output(2, fmt.Sprintf("Field #%d: %s (expected %v)", i, l[i], f))
				panic(err)
			}
			l[i] = parseField(l[i], f)
		}
	}
}

func parseLine(l string) parsedLine {
	p := parsedLine{src: l}
	switch l[0:1] {
	case "1":
		p.kind = "empresa"
		p.contents = [][]string{parseCompany(l)}
		p.validate(CompanySchema)
	case "2":
		p.kind = "socio"
		p.contents = [][]string{parsePartner(l)}
		p.validate(PartnerSchema)
	case "6":
		p.kind = "cnae"
		p.contents = parseCNAE(l)
		p.validate(CNAESchema)
	}
	return p
}
