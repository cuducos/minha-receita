package tools

import (
	"fmt"
	"strings"
)

var validUFs = map[string]bool{
	"AC": true, "AL": true, "AP": true, "AM": true, "BA": true,
	"CE": true, "DF": true, "ES": true, "GO": true, "MA": true,
	"MT": true, "MS": true, "MG": true, "PA": true, "PB": true,
	"PR": true, "PE": true, "PI": true, "RJ": true, "RN": true,
	"RS": true, "RO": true, "RR": true, "SC": true, "SP": true,
	"SE": true, "TO": true,
}

func ValidateUF(uf string) (string, error) {
	uf = strings.ToUpper(strings.TrimSpace(uf))
	if !validUFs[uf] {
		return "", fmt.Errorf("UF inválida: %s", uf)
	}
	return uf, nil
}
