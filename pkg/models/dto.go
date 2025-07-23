package models

import (
	"encoding/json"
	"regexp"
)

type CEPInput struct {
	Cep string `json:"cep"`
}

func NewCEPInput(cep string) *CEPInput {
	return &CEPInput{
		Cep: cep,
	}
}

func (c *CEPInput) GetCep() string {
	return c.Cep
}

func (c *CEPInput) Validar() *RespostaErro {
	if c.Cep == "" {
		return NewRespostaErro(422, "invalid zipcode")
	}
	if len(c.Cep) != 8 {
		return NewRespostaErro(422, "invalid zipcode")
	}

	matched, _ := regexp.MatchString("^[0-9]{8}$", c.Cep)
	if !matched {
		return NewRespostaErro(422, "invalid zipcode")
	}

	return nil
}

func (c *CEPInput) ToStringJson() (*[]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
