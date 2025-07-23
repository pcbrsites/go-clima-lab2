package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pcbrsites/go-clima-lab2/pkg/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type WeatherService struct {
	httpClient *http.Client
	apiKey     string
	tracer     trace.Tracer
}

func NewWeatherService(apiKey string) *WeatherService {
	return &WeatherService{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
		tracer:     otel.Tracer("weather-service"),
	}
}

func (s *WeatherService) BuscarClima(ctx context.Context, cidade string) (*models.WeatherAPI, error) {
	ctx, span := s.tracer.Start(ctx, "buscar_clima")
	defer span.End()

	span.SetAttributes(attribute.String("cidade", cidade))

	baseURL := "https://api.weatherapi.com/v1/current.json"
	params := url.Values{}
	params.Add("key", s.apiKey)
	params.Add("q", cidade)
	params.Add("aqi", "no")

	endpoint := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	log.Println("Buscando clima para a cidade:", cidade)
	log.Println("URL da API:", strings.Replace(endpoint, s.apiKey, "[xxxx]", 1))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("erro na API Weather: status %d", resp.StatusCode)
		return nil, err
	}

	var weather models.WeatherAPI
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.Float64("temperatura_celsius", weather.Current.TempC),
		attribute.String("condicao", weather.Current.Condition.Text),
	)

	return &weather, nil
}
