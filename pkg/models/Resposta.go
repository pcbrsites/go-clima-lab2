package models

type RespostaErro struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RespostaSucesso struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func NewTemperaturaRespostaSucesso(city string, tempC float64) RespostaSucesso {
	tempF := tempC*1.8 + 32
	tempK := tempC + 273

	return RespostaSucesso{
		City:  city,
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}
}

func NewRespostaErro(code int, message string) *RespostaErro {

	return &RespostaErro{
		Code:    code,
		Message: message,
	}
}
