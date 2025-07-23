package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pcbrsites/go-clima-lab2/config"
	"github.com/pcbrsites/go-clima-lab2/internal/service-b/web"
	"github.com/pcbrsites/go-clima-lab2/pkg/telemetry"
)

func main() {
	fmt.Println("o Servidor B está iniciando...")
	cfg := loadConfig()

	tp, err := telemetry.InitTracer("microservice-B", cfg.ZipkinURL)
	if err != nil {
		log.Fatalf("falha ao iniciar tracer: %v", err)
	}

	webServer := web.NewServidorB(cfg.Host, cfg.Porta, cfg.WeatherApiKey)
	webServer.Start()

	fmt.Println("o Servidor B foi iniciado com sucesso.")
	fmt.Println("aguardando conexões...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("desligando o Servidor  B...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	telemetry.ShutdownTracer(ctx, tp)

	fmt.Println("o servidor B foi desligado.")
}

func loadConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("erro ao carregar configuração: %v", err))
	}
	cfg.ShowConfig()

	return cfg
}
