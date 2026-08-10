// Command eventgen is a one-off job: it expands seeds/scenarios.json into
// synthetic `events` rows (the same generator internal/bootstrap uses to
// seed ClickHouse directly at API startup) and publishes them to a Kafka
// topic instead. It is deliberately not wired into cmd/api or the default
// docker-compose build — see clickhouse/kafka/010_events_kafka.sql for the
// Kafka table + materialized view that ingests this topic into
// recap.events, and docker-compose.yml's "events-gen" profile for how to
// run it.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/year-recap/internal/bootstrap"
	"github.com/year-recap/internal/recap/model"
)

// wireEvent is the Kafka payload shape, deliberately distinct from
// model.Event's own JSON tags: its field names match the recap.events
// column names exactly, since the ClickHouse Kafka engine table maps
// JSONEachRow keys to columns by name.
type wireEvent struct {
	ID         string  `json:"id"`
	ProfileID  string  `json:"profile_id"`
	EventType  string  `json:"event_type"`
	OccurredAt string  `json:"occurred_at"`
	Category   string  `json:"category"`
	AdID       *uint64 `json:"ad_id"`
	DialogID   *uint64 `json:"dialog_id"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	brokers := splitCSV(envOrDefault("EVENTGEN_KAFKA_BROKERS", "kafka:9092"))
	topic := envOrDefault("EVENTGEN_KAFKA_TOPIC", "events")
	profilesPath := envOrDefault("PROFILES_PATH", "seeds/profiles.json")
	scenariosPath := envOrDefault("SCENARIOS_PATH", "seeds/scenarios.json")

	data, err := bootstrap.GenerateDemoData(profilesPath, scenariosPath)
	if err != nil {
		return fmt.Errorf("generate demo events: %w", err)
	}
	if len(data.Events) == 0 {
		log.Println("no events generated from scenarios, nothing to publish")
		return nil
	}

	// The topic may not exist yet on a fresh broker: create it up front
	// instead of relying on Writer.AllowAutoTopicCreation, whose first
	// produce attempt otherwise races the broker's own auto-creation and can
	// fail with "Unknown Topic Or Partition".
	if err := ensureTopic(brokers[0], topic); err != nil {
		return fmt.Errorf("ensure kafka topic %q: %w", topic, err)
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		RequiredAcks:           kafka.RequireOne,
	}
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("close kafka writer: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := make([]kafka.Message, 0, len(data.Events))
	for _, event := range data.Events {
		payload, err := json.Marshal(toWireEvent(event))
		if err != nil {
			return fmt.Errorf("encode event %s: %w", event.ID, err)
		}
		messages = append(messages, kafka.Message{
			Key:   []byte(event.ProfileID.String()),
			Value: payload,
		})
	}

	if err := writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("publish events to kafka: %w", err)
	}

	log.Printf("published %d events for %d profiles to topic %q on %s", len(data.Events), len(data.Profiles), topic, strings.Join(brokers, ","))
	return nil
}

func ensureTopic(broker, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("dial %s: %w", broker, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("close kafka conn: %v", err)
		}
	}()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("find controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("dial controller: %w", err)
	}
	defer func() {
		if err := controllerConn.Close(); err != nil {
			log.Printf("close kafka controller conn: %v", err)
		}
	}()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
		ConfigEntries: []kafka.ConfigEntry{
			// Explicit, not the broker's default 7-day retention: this
			// topic is transport into recap.events (see
			// clickhouse/kafka/010_events_kafka.sql), not itself the
			// source of truth, so it doesn't need to hold messages for
			// long — just long enough to survive the consumer (the
			// ClickHouse Kafka engine table) being briefly down or not
			// started yet. 24h is a generous buffer for an on-demand demo
			// job, not a number backed by an incident/runbook.
			{ConfigName: "retention.ms", ConfigValue: strconv.Itoa(int(24 * time.Hour / time.Millisecond))},
		},
	})
	if err != nil && !errors.Is(err, kafka.TopicAlreadyExists) {
		return err
	}

	// CreateTopics returning success only means the controller accepted the
	// request — the topic's metadata (partition leader assignment) still
	// needs to propagate to the broker(s) before a produce request against
	// it will succeed. On a freshly created topic that gap is real, if
	// short: without waiting it out here, the very next WriteMessages call
	// intermittently fails with "Unknown Topic Or Partition".
	return waitForTopic(conn, topic, 10, 500*time.Millisecond)
}

func waitForTopic(conn *kafka.Conn, topic string, attempts int, backoff time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		partitions, err := conn.ReadPartitions(topic)
		if err == nil && len(partitions) > 0 {
			return nil
		}
		lastErr = err
		time.Sleep(backoff)
	}
	return fmt.Errorf("topic %q not ready after %d attempts: %w", topic, attempts, lastErr)
}

func toWireEvent(event model.Event) wireEvent {
	return wireEvent{
		ID:         event.ID.String(),
		ProfileID:  event.ProfileID.String(),
		EventType:  string(event.Type),
		OccurredAt: event.OccurredAt.UTC().Format("2006-01-02 15:04:05"),
		Category:   event.Category,
		AdID:       event.AdID,
		DialogID:   event.DialogID,
	}
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(raw string) []string {
	var values []string
	for _, item := range strings.Split(raw, ",") {
		if value := strings.TrimSpace(item); value != "" {
			values = append(values, value)
		}
	}
	return values
}
