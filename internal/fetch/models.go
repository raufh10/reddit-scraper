package fetch

import "time"

// Targets holds the URLs and subreddits for the scraper.
type Targets struct {
  BaseURL    string
  Subreddits []string
}

// ScraperConfigs holds the technical settings for the scraper instance.
type ScraperConfigs struct {
  UserAgent      string
  ProxyURL       string
  TimeoutSeconds time.Duration
}

type Client struct {
  Targets Targets
  Configs []ScraperConfigs
}
