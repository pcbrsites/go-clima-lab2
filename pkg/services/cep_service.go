package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pcbrsites/go-clima-lab2/pkg/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type CEPService struct {
	httpClient *http.Client
	tracer     trace.Tracer
}

func NewCEPService() *CEPService {
	return &CEPService{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		tracer:     otel.Tracer("cep-service"),
	}
}

func (s *CEPService) BuscarCEP(ctx context.Context, cep string) (*models.ViaCep, error) {
	ctx, span := s.tracer.Start(ctx, "buscar_cep")
	defer span.End()

	span.SetAttributes(attribute.String("cep", cep))

	log.Println("Buscando CEP:", cep)

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {

		return nil, err
	}

	resp, err := s.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("erro na API ViaCEP: status %d", resp.StatusCode)
		log.Println(err)

		return nil, err
	}

	var viaCep models.ViaCep
	if err := json.NewDecoder(resp.Body).Decode(&viaCep); err != nil {
		log.Println(err)
		return nil, err
	}

	if viaCep.Erro != "" {
		err := fmt.Errorf("CEP não encontrado")
		log.Println(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("cidade", viaCep.Localidade),
		attribute.String("uf", viaCep.Uf),
	)

	return &viaCep, nil
}
