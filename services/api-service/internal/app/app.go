package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"mephi_vkr_aspm/services/api-service/internal/agentdebug"
	"mephi_vkr_aspm/services/api-service/internal/config"
	"mephi_vkr_aspm/services/api-service/internal/httpapi"
	pkgkafka "mephi_vkr_aspm/services/api-service/internal/kafka"
	"mephi_vkr_aspm/services/api-service/internal/service"
)

type App struct {
	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	var kafkaBridge *pkgkafka.IngestBridge
	if len(cfg.KafkaBrokers) > 0 {
		// #region agent log
		agentdebug.Log("H1", "internal/app/app.go:New", "before EnsureTopics", map[string]any{
			"brokers":       cfg.KafkaBrokers,
			"ingestTopic":   cfg.KafkaTopicIngest,
			"resultTopic":   cfg.KafkaTopicResult,
			"kafkaBrokerN": len(cfg.KafkaBrokers),
		})
		// #endregion
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := pkgkafka.EnsureTopics(ctx, cfg.KafkaBrokers, cfg.KafkaTopicIngest, cfg.KafkaTopicResult)
		cancel()
		if err != nil {
			// #region agent log
			agentdebug.Log("H1", "internal/app/app.go:New", "EnsureTopics failed", map[string]any{"error": err.Error()})
			// #endregion
			return nil, fmt.Errorf("kafka ensure topics: %w", err)
		}
		// #region agent log
		agentdebug.Log("H1", "internal/app/app.go:New", "EnsureTopics ok", nil)
		// #endregion
		kafkaBridge = pkgkafka.NewIngestBridge(cfg.KafkaBrokers, cfg.KafkaTopicIngest, cfg.KafkaTopicResult)
		log.Printf("api-service: findings ingest via Kafka (%s -> %s)", cfg.KafkaTopicIngest, cfg.KafkaTopicResult)
	} else {
		log.Printf("api-service: findings ingest via HTTP (APP_KAFKA_BROKERS empty)")
	}

	orchestrator := service.New(
		cfg.ProcessingServiceURL,
		cfg.JiraServiceURL,
		cfg.SemgrepServiceURL,
		kafkaBridge,
	)

	mux := http.NewServeMux()
	handler := httpapi.New(orchestrator)
	handler.Register(mux)

	return &App{
		server: &http.Server{
			Addr:    ":" + cfg.HTTPPort,
			Handler: mux,
		},
	}, nil
}

func (a *App) Run() error {
	log.Printf("api-service listening on %s", a.server.Addr)
	return a.server.ListenAndServe()
}
