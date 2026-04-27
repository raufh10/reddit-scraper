package parser

// RedditPostData represents the specific fields extracted during parsing
type RedditPostData struct {
  RedditID          string                   `json:"name"`
  Subreddit         string                   `json:"subreddit"`
  Title             string                   `json:"title"`
  Selftext          string                   `json:"selftext"`
  URL               string                   `json:"url"`
  CreatedUTC        float64                  `json:"created_utc"`
  LinkFlairRichText []map[string]interface{} `json:"link_flair_richtext"`
}

// RedditPostChild wraps the individual post data
type RedditPostChild struct {
  Data RedditPostData `json:"data"`
}

// RedditData represents the "data" block of the Reddit API response
type RedditData struct {
  Children []RedditPostChild `json:"children"`
}

// RedditResponse is the entry point for the JSON unmarshaling
type RedditResponse struct {
  Data RedditData `json:"data"`
}

