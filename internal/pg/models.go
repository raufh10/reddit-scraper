package pg

import (
  "encoding/json"
  "time"

  "github.com/google/uuid"
)

// StorablePost represents the raw data coming from the scraper
type StorablePost struct {
  RedditID    string                 `json:"reddit_id"`
  Subreddit   string                 `json:"subreddit"`
  Title       string                 `json:"title"`
  Content     string                 `json:"content"`
  URL         string                 `json:"url"`
  Ups         int                    `json:"ups"`          
  UpvoteRatio float64                `json:"upvote_ratio"` 
  PostedAt    float64                `json:"posted_at"`
  Metadata    map[string]interface{} `json:"metadata"`
}

// NewsPost maps directly to your public.news_posts schema for reading
type NewsPost struct {
  ID          uuid.UUID       `db:"id"`
  RedditID    string          `db:"reddit_id"`
  Subreddit   string          `db:"subreddit"`
  Title       string          `db:"title"`
  Content     *string         `db:"content"`
  URL         string          `db:"url"`         
  Ups         *int            `db:"ups"`
  UpvoteRatio *float64        `db:"upvote_ratio"`
  PostedAt    *time.Time      `db:"posted_at"`
  CreatedAt   time.Time       `db:"created_at"`
  IsPosted    *bool           `db:"is_posted"`
  Metadata    json.RawMessage `db:"metadata"`
}

