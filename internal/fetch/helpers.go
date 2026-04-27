package fetch

import (
  "encoding/json"
  "fmt"
  "math/rand"
  "net/http"
  "net/url"
  "os"
  "time"
)

// fetchFreshUserAgents retrieves a list of UAs from ScrapeOps.
func fetchFreshUserAgents(apiKey string) ([]string, error) {
  apiURL := fmt.Sprintf("http://headers.scrapeops.io/v1/user-agents?api_key=%s&num_results=5", apiKey)

  resp, err := http.Get(apiURL)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("scrapeops api status: %d", resp.StatusCode)
  }

  var data struct {
    Result []string `json:"result"`
  }

  if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
    return nil, err
  }

  return data.Result, nil
}

// generateOxylabsProxy builds the authenticated proxy string.
func generateOxylabsProxy() (string, error) {
  user := os.Getenv("OXYLABS_USER")
  key := os.Getenv("OXYLABS_KEY")

  if user == "" || key == "" {
    return "", fmt.Errorf("OXYLABS credentials not set")
  }

  username := url.QueryEscape(user)
  password := url.QueryEscape(key)
  sessionID := rand.Intn(999999-100000) + 100000

  fullUsername := fmt.Sprintf("customer-%s-sessid-%d", username, sessionID)

  return fmt.Sprintf("http://%s:%s@pr.oxylabs.io:7777", fullUsername, password), nil
}

// GetBackoffDuration returns 10s, 20s, or 30s based on attempt number
func GetBackoffDuration(attempt int) time.Duration {
  switch attempt {
  case 1:
    return 10 * time.Second
  case 2:
    return 20 * time.Second
  case 3:
    return 30 * time.Second
  default:
    return 10 * time.Second
  }
}

// GetWorkerDelay returns a randomized duration within a specific slot.
// For example, if there are 10 workers, each gets a 1-minute window:
// Worker 0: 0-60s, Worker 1: 61-120s, etc.
func GetWorkerDelay(workerIndex, totalWorkers int) time.Duration {
  if totalWorkers <= 0 {
    return 0
  }

  const totalWindow = 10 * time.Minute
  slotDuration := totalWindow / time.Duration(totalWorkers)
  
  // Calculate the start of the worker's slot
  slotStart := time.Duration(workerIndex) * slotDuration
  
  // Generate a random offset within that slot
  randomOffset := time.Duration(rand.Int63n(int64(slotDuration)))
  
  return slotStart + randomOffset
}
