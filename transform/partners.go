package transform

import (
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

type partner struct {
	IdentificadorDeSocio                 *int    `json:"identificador_de_socio"`
	NomeSocio                            string  `json:"nome_socio"`
	CNPJCPFDoSocio                       string  `json:"cnpj_cpf_do_socio"`
	CodigoQualificacaoSocio              *int    `json:"codigo_qualificacao_socio"`
	QualificaoSocio                      *string `json:"qualificacao_socio"`
	DataEntradaSociedade                 *date   `json:"data_entrada_sociedade"`
	CodigoPais                           *int    `json:"codigo_pais"`
	Pais                                 *string `json:"pais"`
	CPFRepresentanteLegal                string  `json:"cpf_representante_legal"`
	NomeRepresentanteLegal               string  `json:"nome_representante_legal"`
	CodigoQualificacaoRepresentanteLegal *int    `json:"codigo_qualificacao_representante_legal"`
	QualificacaoRepresentanteLegal       *string `json:"qualificacao_representante_legal"`
	CodigoFaixaEtaria                    *int    `json:"codigo_faixa_etaria"`
	FaixaEtaria                          *string `json:"faixa_etaria"`
}

func (p *partner) faixaEtaria(v string) {
	c, err := toInt(v)
	if err != nil {
		return
	}
	p.CodigoFaixaEtaria = c

	var s string
	switch *c {
	case 1:
		s = "para os intervalos entre 0 a 12 anos"
	case 2:
		s = "Entre 13 a 20 ano"
	case 3:
		s = "Entre 21 a 30 anos"
	case 4:
		s = "Entre 31 a 40 anos"
	case 5:
		s = "Entre 41 a 50 anos"
	case 6:
		s = "Entre 51 a 60 anos"
	case 7:
		s = "Entre 61 a 70 anos"
	case 8:
		s = "Entre 71 a 80 anos"
	case 9:
		s = "Maiores de 80 anos"
	case 0:
		s = "Não se aplica"
	}
	if s != "" {
		p.FaixaEtaria = &s
	}
	return
}

func (p *partner) pais(l *lookups, v string) error {
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoPais %s: %w", v, err)
	}
	if i == nil {
		return nil
	}
	s := l.countries[*i]
	p.CodigoPais = i
	if s != "" {
		p.Pais = &s
	}
	return nil
}

func newPartner(l *lookups, r []string) (partner, error) {
	identificadorDeSocio, err := toInt(r[1])
	if err != nil {
		return partner{}, fmt.Errorf("error parsing IdentificadorDeSocio %s: %w", r[1], err)
	}

	codigoQualificacaoSocio, err := toInt(r[4])
	if err != nil {
		return partner{}, fmt.Errorf("error parsing CodigoQualificacaoSocio %s: %w", r[4], err)
	}

	dataEntradaSociedade, err := toDate(r[5])
	if err != nil {
		return partner{}, fmt.Errorf("error parsing DataEntradaSociedade %s: %w", r[5], err)
	}

	codigoQualificacaoRepresentanteLegal, err := toInt(r[9])
	if err != nil {
		return partner{}, fmt.Errorf("error parsing CodigoQualificacaoRepresentanteLegal %s: %w", r[9], err)
	}

	p := partner{
		IdentificadorDeSocio:                 identificadorDeSocio,
		NomeSocio:                            r[2],
		CNPJCPFDoSocio:                       r[3],
		CodigoQualificacaoSocio:              codigoQualificacaoSocio,
		DataEntradaSociedade:                 dataEntradaSociedade,
		CPFRepresentanteLegal:                r[7],
		NomeRepresentanteLegal:               r[8],
		CodigoQualificacaoRepresentanteLegal: codigoQualificacaoRepresentanteLegal,
	}
	p.pais(l, r[6])
	p.faixaEtaria(r[10])
	return p, nil
}

type partnersTask struct {
	dir        string
	lookups    *lookups
	success    chan struct{}
	errors     chan error
	bar        *progressbar.ProgressBar
	filesMutex mapMutex
}

func (t *partnersTask) addPartner(r []string) {
	p, err := newPartner(t.lookups, r)
	if err != nil {
		t.errors <- fmt.Errorf("error creating partner for %v: %w", r, err)
		return
	}
	ls, err := filepath.Glob(filepath.Join(t.dir, r[0], "*.json"))
	if err != nil {
		t.errors <- fmt.Errorf("error in the glob pattern: %w", err)
		return
	}
	if len(ls) == 0 {
		log.Output(2, fmt.Sprintf("No JSON file found for CNPJ base %s", r[0]))
		t.bar.Add(1)
		return
	}
	for _, f := range ls {
		go func(t *partnersTask, f string, p partner) {
			t.filesMutex.lock(f)
			defer t.filesMutex.unlock(f)

			c, err := companyFromJSON(f)
			if err != nil {
				t.errors <- fmt.Errorf("error reading company from %s: %w", f, err)
			}
			c.QuadroSocietario = append(c.QuadroSocietario, p)
			f, err = c.toJSON(t.dir)
			if err != nil {
				t.errors <- err
				return
			}
			t.success <- struct{}{}
		}(t, f, p)
	}
	return
}

func addPartners(dir string, l *lookups) error {
	s, err := newSource(partners, dir)
	if err != nil {
		return fmt.Errorf("error creating source for partners: %w", err)
	}
	defer s.close()

	t := partnersTask{
		dir:        dir,
		lookups:    l,
		success:    make(chan struct{}),
		errors:     make(chan error),
		bar:        progressbar.Default(s.totalLines),
		filesMutex: newMapMutex(),
	}
	for _, r := range s.readers {
		go func(t *partnersTask, a *archivedCSV) {
			for {
				r, err := a.read()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.errors <- err
					break // do not proceed in case of errors.
				}
				go t.addPartner(r)
			}
		}(&t, r)
	}

	defer func() {
		close(t.success)
		close(t.errors)
	}()

	for {
		select {
		case err := <-t.errors:
			fmt.Println(err)
			return err
		case <-t.success:
			t.bar.Add(1)
			if t.bar.IsFinished() {
				return nil
			}
		}
	}
}
