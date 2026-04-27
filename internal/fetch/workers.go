package fetch

import (
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "net/url"
  "sync"
  "time"

  pr "scraper/internal/parser"
  pg "scraper/internal/pg"
)

// Run starts concurrent workers for each target subreddit and returns a flattened result set.
func (c *Client) Run(limit int, pages int) []pg.StorablePost {
  var wg sync.WaitGroup
  numSubreddits := len(c.Targets.Subreddits)
  
  // Mutex to protect the shared slice during concurrent appends
  var mu sync.Mutex
  finalResults := make([]pg.StorablePost, 0)

  for i, subreddit := range c.Targets.Subreddits {
    wg.Add(1)

    idx := i
    sub := subreddit
    config := c.Configs[idx%len(c.Configs)]

    go func(workerIdx int, targetSub string, cfg ScraperConfigs) {
      defer wg.Done()
      
      // Recovery from unexpected panics in a single worker (e.g., malformed JSON)
      defer func() {
        if r := recover(); r != nil {
          log.Printf("[Worker %d] Recovered from panic while processing %s: %v", workerIdx, targetSub, r)
        }
      }()

      // 1. Staggered start delay within the 10-minute range
      delay := GetWorkerDelay(workerIdx, numSubreddits)
      time.Sleep(delay)

      httpClient, err := c.InitHttpClient(cfg)
      if err != nil {
        log.Printf("[Worker %d] Client Init Error for %s: %v", workerIdx, targetSub, err)
        return
      }

      var after string
      for p := 0; p < pages; p++ {
        var targetURL string
        if after == "" {
          targetURL = c.GetSubredditURL(targetSub, limit)
        } else {
          targetURL = c.GetSubredditPaginationURL(targetSub, after, limit)
        }

        // 2. Retry Logic with Backoff (10s, 20s, 30s)
        var resp *pr.RedditResponse
        var fetchErr error
        for attempt := 1; attempt <= 3; attempt++ {
          resp, fetchErr = c.FetchSubredditJson(httpClient, targetURL, cfg.UserAgent)
          if fetchErr == nil {
            break
          }
          log.Printf("[Worker %d] Attempt %d failed for %s: %v", workerIdx, attempt, targetSub, fetchErr)
          time.Sleep(GetBackoffDuration(attempt))
        }

        // If a specific subreddit fails after all retries, exit this worker gracefully
        if fetchErr != nil {
          log.Printf("[Worker %d] Abandoning %s after 3 failed attempts", workerIdx, targetSub)
          break
        }

        // 3. Parse and Append safely using Mutex
        pagePosts := pr.ProcessResponse(*resp)
        
        mu.Lock()
        finalResults = append(finalResults, pagePosts...)
        mu.Unlock()

        // 4. Pagination Check
        if resp.Data.After == nil || *resp.Data.After == "" {
          break
        }
        after = *resp.Data.After

        // Delay between page fetches to respect API limits
        time.Sleep(2 * time.Second)
      }
    }(idx, sub, config)
  }

  // Wait for all goroutines to finish their assigned work or retry attempts
  wg.Wait()
  return finalResults
}

// InitHttpClient initializes a client using a specific ScraperConfig
func (c *Client) InitHttpClient(config ScraperConfigs) (*http.Client, error) {
  transport := &http.Transport{
    MaxIdleConns:          10,
    IdleConnTimeout:       90 * time.Second,
    TLSHandshakeTimeout:   10 * time.Second,
    ExpectContinueTimeout: 1 * time.Second,
  }

  if config.ProxyURL != "" {
    proxyURL, err := url.Parse(config.ProxyURL)
    if err != nil {
      return nil, fmt.Errorf("invalid proxy URL: %w", err)
    }
    transport.Proxy = http.ProxyURL(proxyURL)
  }

  return &http.Client{
    Transport: transport,
    Timeout:   config.TimeoutSeconds,
  }, nil
}

// FetchSubredditJson performs the actual network request
func (c *Client) FetchSubredditJson(httpClient *http.Client, targetURL string, ua string) (*pr.RedditResponse, error) {
  req, err := http.NewRequest("GET", targetURL, nil)
  if err != nil {
    return nil, err
  }

  req.Header.Set("User-Agent", ua)
  req.Header.Set("Accept", "application/json")

  resp, err := httpClient.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("reddit error status: %d", resp.StatusCode)
  }

  var result pr.RedditResponse
  if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
    return nil, err
  }

  return &result, nil
}

