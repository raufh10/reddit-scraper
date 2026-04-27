package fetch

import (
  "fmt"
  "os"
  "time"

  "github.com/joho/godotenv"
)

func NewConfig(subreddits []string) (*Client, error) {
  _ = godotenv.Load()

  scrapeOpsKey := os.Getenv("SCRAPEOPS_API_KEY")
  if scrapeOpsKey == "" {
    return nil, fmt.Errorf("SCRAPEOPS_API_KEY must be set")
  }

  // Fetch a list of user agents (one for each subreddit/config)
  agents, err := fetchFreshUserAgents(scrapeOpsKey)
  if err != nil {
    return nil, fmt.Errorf("failed to fetch user agents: %w", err)
  }

  if len(agents) == 0 {
    return nil, fmt.Errorf("no user agents returned from provider")
  }

  var configs []ScraperConfigs
  
  // Generate a ScraperConfig for each subreddit in the list
  for i := 0; i < len(subreddits); i++ {
    agent := agents[i%len(agents)]
    proxy, err := generateOxylabsProxy()
    if err != nil {
      continue
    }

    configs = append(configs, ScraperConfigs{
      UserAgent:      agent,
      ProxyURL:       proxy,
      TimeoutSeconds: 120 * time.Second,
    })
  }

  return &Client{
    Targets: Targets{
      BaseURL:    "https://www.reddit.com",
      Subreddits: subreddits,
    },
    Configs: configs,
  }, nil
}

// GetSubredditURL assembles target SubredditURL with limit int
func (c *Client) GetSubredditURL(subreddit string, limit int) string {
  return fmt.Sprintf("%s/r/%s.json?limit=%d", c.Targets.BaseURL, subreddit, limit)
}

// GetSubredditPaginationURL assembles target SubredditPaginationURL with extracted after string & limit int
func (c *Client) GetSubredditPaginationURL(subreddit, after string, limit int) string {
  return fmt.Sprintf("%s/r/%s.json?limit=%d&after=%s", c.Targets.BaseURL, subreddit, limit, after)
}
