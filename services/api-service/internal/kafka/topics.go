package kafka

import (
	"context"
	"errors"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"

	"mephi_vkr_aspm/services/api-service/internal/agentdebug"
)

// EnsureTopics создаёт топики (идемпотентно).
func EnsureTopics(ctx context.Context, brokers []string, topicNames ...string) error {
	if len(brokers) == 0 {
		return errors.New("no kafka brokers")
	}
	addr := kafkago.TCP(brokers[0])
	cli := &kafkago.Client{Addr: addr}
	topics := make([]kafkago.TopicConfig, len(topicNames))
	for i, name := range topicNames {
		topics[i] = kafkago.TopicConfig{
			Topic:             name,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
	}
	resp, err := cli.CreateTopics(ctx, &kafkago.CreateTopicsRequest{
		Addr:   addr,
		Topics: topics,
	})
	if err != nil {
		// #region agent log
		agentdebug.Log("H1", "internal/kafka/topics.go:EnsureTopics", "CreateTopics transport error", map[string]any{
			"broker": brokers[0],
			"error":  err.Error(),
		})
		// #endregion
		return fmt.Errorf("create topics: %w", err)
	}
	for name, e := range resp.Errors {
		if e == nil {
			continue
		}
		if errors.Is(e, kafkago.TopicAlreadyExists) {
			continue
		}
		if name != "" {
			// #region agent log
			agentdebug.Log("H3", "internal/kafka/topics.go:EnsureTopics", "CreateTopics response error", map[string]any{
				"topic": name,
				"error": e.Error(),
			})
			// #endregion
			return fmt.Errorf("topic %q: %w", name, e)
		}
	}
	return nil
}
