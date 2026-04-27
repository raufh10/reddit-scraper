package nats

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/nats-io/nats.go"
)

// ScraperService is what your scraper package will implement
type ScraperService interface {
	Run(event ScraperEvent)
}

type NatsWorker struct {
	client  *Client
	scraper ScraperService
}

func NewNatsWorker(nc *Client, scraper ScraperService) *NatsWorker {
	return &NatsWorker{
		client:  nc,
		scraper: scraper,
	}
}

// Start listens to the subject defined in your EventConfig
func (w *NatsWorker) Start(subject string) error {
	_, err := w.client.Listen(subject, w.handleScraperMessage)
	return err
}

func (w *NatsWorker) handleScraperMessage(m *nats.Msg) {
	var rawPayload ScraperPayload
	if err := json.Unmarshal(m.Data, &rawPayload); err != nil {
		log.Printf("[!] JSON Unmarshal error: %v", err)
		return
	}

	// Transform "sub1,sub2" into []string{"sub1", "sub2"}
	targets := strings.Split(rawPayload.Target, ",")
	for i := range targets {
		targets[i] = strings.TrimSpace(targets[i])
	}

	event := ScraperEvent{
		Targets: targets,
		Pages:   rawPayload.Pages,
		Limit:   rawPayload.Limit,
	}

	log.Printf("[*] Received targets: %v. Starting concurrent jobs...", event.Targets)
	
	// Hand off to the actual scraper logic
	go w.scraper.Run(event)
}

