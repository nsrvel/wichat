package kafka

import "strings"

// Config wires Kafka producers and consumers.
type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

// ParseBrokers splits a comma-separated broker list.
func ParseBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
