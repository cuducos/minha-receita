package transform

import "fmt"

type lookup map[int]string

type lookups struct {
	motives lookup
}

func newLookups() lookups {
	return lookups{motives: make(map[int]string)}
}

func (c *company) motivoSituacaoCadastral(l lookups, v string) error {
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse MotivoSituacaoCadastral %s: %w", v, err)
	}

	s := l.motives[*i]
	c.MotivoSituacaoCadastral = i
	if s != "" {
		c.DescricaoMotivoSituacaoCadastral = &s
	}
	return nil
}
