package parser

import (
  "encoding/json"

  pg "scraper/internal/pg"
)

// ProcessResponse extracts StorablePost data from RedditResponse
func ProcessResponse(resp RedditResponse) []pg.StorablePost {
  posts := make([]pg.StorablePost, 0, len(resp.Data.Children))

  for _, child := range resp.Data.Children {
    data := child.Data

    var metadata map[string]interface{}
    rawBytes, err := json.Marshal(data)
    if err == nil {
      _ = json.Unmarshal(rawBytes, &metadata)
    }

    posts = append(posts, pg.StorablePost{
      RedditID:  data.RedditID,
      Subreddit: data.Subreddit,
      Title:     data.Title,
      Content:   data.Selftext,
      URL:       data.URL,
      PostedAt:  data.CreatedUTC,
      Metadata:  metadata,
    })
  }

  return posts
}

