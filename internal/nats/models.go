package nats

// ScraperPayload represents the raw JSON structure coming from NATS
type ScraperPayload struct {
  Target string `json:"target"`
  Pages  int    `json:"pages"`
  Limit  int    `json:"limit"`
}

// ScraperEvent is the cleaned, domain-ready version for your scraper
type ScraperEvent struct {
  Targets []string // Cleaned slice: ["golang", "rust", "zig"]
  Pages   int
  Limit   int
}

// EventConfig handles your YAML configuration mapping
type EventConfig struct {
  Scraper struct {
    Cron    string `yaml:"cron"`
    Subject string `yaml:"subject"`
    Payload any    `yaml:"payload"`
  } `yaml:"scraper"`
}

