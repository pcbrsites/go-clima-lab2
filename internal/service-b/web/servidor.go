package web

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pcbrsites/go-clima-lab2/pkg/models"
	"github.com/pcbrsites/go-clima-lab2/pkg/services"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ServidorB struct {
	Host           string
	Porta          string
	weatherService *services.WeatherService
	cepService     *services.CEPService
	tracer         trace.Tracer
}

func NewServidorB(host string, porta string, weatherApiKey string) *ServidorB {
	return &ServidorB{
		Host:           host,
		Porta:          porta,
		weatherService: services.NewWeatherService(weatherApiKey),
		cepService:     services.NewCEPService(),
		tracer:         otel.Tracer("service-b"),
	}
}

func (s *ServidorB) Start() {

	r := gin.New()

	r.Use(otelgin.Middleware("service-b"))

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.POST("/", s.handleProcessarCEP)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "service"})
	})

	addr := fmt.Sprintf("%s:%s", s.Host, s.Porta)
	log.Printf("Servidor Service B rodando em %s\n", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			panic(fmt.Sprintf("Erro ao iniciar servidor: %v", err))
		}
	}()
}

func (s *ServidorB) handleProcessarCEP(c *gin.Context) {

	for key, values := range c.Request.Header {
		for _, value := range values {
			log.Printf("Header recebido: %s=%s", key, value)
		}
	}

	ctx := c.Request.Context()
	log.Printf("TraceID extraído no Serviço B: %s", trace.SpanContextFromContext(ctx).TraceID())

	ctx, span := s.tracer.Start(ctx, "service-b")
	defer span.End()

	var cepInput models.CEPInput
	if err := c.ShouldBindJSON(&cepInput); err != nil {

		c.JSON(422, gin.H{"message": "invalid zipcode"})
		return
	}

	span.SetAttributes(attribute.String("cep", cepInput.Cep))

	if erro := cepInput.Validar(); erro != nil {
		c.JSON(erro.Code, gin.H{"message": erro.Message})
		return
	}

	// Buscar localização pelo CEP
	viaCep, err := s.buscarLocalizacao(ctx, cepInput.Cep)
	if err != nil {
		c.JSON(404, gin.H{"message": "can not find zipcode"})
		return
	}

	// Buscar clima
	clima, err := s.buscarClima(ctx, viaCep.Localidade)
	if err != nil {

		c.JSON(500, gin.H{"message": "error fetching weather data"})
		return
	}

	resposta := models.NewTemperaturaRespostaSucesso(viaCep.Localidade, clima.Current.TempC)

	span.SetAttributes(
		attribute.String("cidade", resposta.City),
		attribute.Float64("temp_celsius", resposta.TempC),
		attribute.Float64("temp_fahrenheit", resposta.TempF),
		attribute.Float64("temp_kelvin", resposta.TempK),
	)

	c.JSON(200, resposta)
}

func (s *ServidorB) buscarLocalizacao(ctx context.Context, cep string) (*models.ViaCep, error) {

	return s.cepService.BuscarCEP(ctx, cep)
}

func (s *ServidorB) buscarClima(ctx context.Context, cidade string) (*models.WeatherAPI, error) {

	return s.weatherService.BuscarClima(ctx, cidade)
}
