package web

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pcbrsites/go-clima-lab2/pkg/models"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Servidor struct {
	Host        string
	Porta       string
	ServiceBURL string
	tracer      trace.Tracer
}

func NewServidor(host string, porta string, serviceBURL string) *Servidor {
	return &Servidor{
		Host:        host,
		Porta:       porta,
		ServiceBURL: serviceBURL,
		tracer:      otel.Tracer("service-a"),
	}
}

func (s *Servidor) Start() {

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("service-a"))

	r.POST("/", s.handleValidarCEP)
	r.POST("/cep", s.handleValidarCEP)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up", "service": "service-a"})
	})

	addr := fmt.Sprintf("%s:%s", s.Host, s.Porta)
	fmt.Printf("servidor Service A rodando em %s\n", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			panic(fmt.Sprintf("erro ao iniciar servidor: %v", err))
		}
	}()
}

func (s *Servidor) handleValidarCEP(c *gin.Context) {
	ctx := c.Request.Context()

	var cepInput models.CEPInput
	if err := c.ShouldBindJSON(&cepInput); err != nil {
		c.JSON(422, gin.H{"message": "invalid zipcode"})
		return
	}

	if erro := cepInput.Validar(); erro != nil {
		c.JSON(erro.Code, gin.H{"message": erro.Message})
		return
	}

	resposta, statusCode, err := s.buscarTemperatura(ctx, cepInput)
	if err != nil {
		c.JSON(500, gin.H{"message": "Internal server error"})
		return
	}

	c.Data(statusCode, "application/json", resposta)
}

func (s *Servidor) buscarTemperatura(ctx context.Context, cepInput models.CEPInput) ([]byte, int, error) {

	jsonData, err := cepInput.ToStringJson()
	if err != nil {
		return nil, 0, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", s.ServiceBURL, bytes.NewBuffer(*jsonData))
	if err != nil {

		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	resposta := buf.Bytes()

	return resposta, resp.StatusCode, nil
}
