package pg

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/jmoiron/sqlx"
  "github.com/lib/pq"
)

// BulkIngestRawPosts handles upserts using Postgres UNNEST.
func BulkIngestRawPosts(ctx context.Context, db *sqlx.DB, posts []StorablePost) error {
  if len(posts) == 0 {
    return nil
  }

  // Pre-allocate slices for performance
  redditIDs := make([]string, len(posts))
  subreddit := make([]string, len(posts))
  titles := make([]string, len(posts))
  contents := make([]string, len(posts))
  urls := make([]string, len(posts))
  postedAts := make([]time.Time, len(posts))
  metadatas := make([]string, len(posts))

  for i, p := range posts {
    redditIDs[i] = p.RedditID
    subreddit[i] = p.Subreddit
    titles[i] = p.Title
    contents[i] = p.Content
    urls[i] = p.URL
    postedAts[i] = time.Unix(int64(p.PostedAt), 0).UTC()

    // Marshal the map into a JSON string for the DB
    metaBytes, err := json.Marshal(p.Metadata)
    if err != nil {
      metadatas[i] = "{}"
    } else {
      metadatas[i] = string(metaBytes)
    }
  }

  const query = `
    INSERT INTO news_posts (reddit_id, subreddit, title, content, url, posted_at, metadata)
    SELECT * FROM UNNEST(
      $1::text[], 
      $2::text[], 
      $3::text[], 
      $4::text[], 
      $5::text[], 
      $6::timestamptz[], 
      $7::jsonb[]
    )
    ON CONFLICT (reddit_id, subreddit) DO UPDATE SET
      title = EXCLUDED.title,
      content = EXCLUDED.content,
      metadata = EXCLUDED.metadata;
  `

  _, err := db.ExecContext(
    ctx,
    query,
    pq.Array(redditIDs),
    pq.Array(subreddit),
    pq.Array(titles),
    pq.Array(contents),
    pq.Array(urls),
    pq.Array(postedAts),
    pq.Array(metadatas),
  )

  if err != nil {
    return fmt.Errorf("bulk ingest failed: %w", err)
  }

  return nil
}
