package main

import (
  "context"
  "log"
  "os"

  "github.com/joho/godotenv"

  ft "scraper/internal/fetch"
  nt "scraper/internal/nats"
  db "scraper/internal/pg"
)

// App holds the logic for the scraper service
type App struct{}

// Run implements nt.ScraperService interface
func (a *App) Run(event nt.ScraperEvent) {
  log.Printf("[*] Triggered: Crawling %v", event.Targets)

  // 1. Setup Database Connection (Option 1: Connection per Event)
  dbURL := os.Getenv("DATABASE_URL")
  if dbURL == "" {
    log.Printf("[-] DATABASE_URL not set. Skipping job.")
    return
  }

  conn, err := db.Connect(dbURL)
  if err != nil {
    log.Printf("[-] Database connection error: %v", err)
    return
  }
  defer conn.Close() // Resources released as soon as Run() finishes

  // 2. Initialize Fetch Client with NATS targets
  client, err := ft.NewConfig(event.Targets)
  if err != nil {
    log.Printf("[-] Failed to initialize fetch client: %v", err)
    return
  }

  // 3. Execute Concurrent Scraper
  posts := client.Run(event.Limit, event.Pages)

  if len(posts) == 0 {
    log.Printf("[!] No posts collected. DB untouched.")
    return
  }

  // 4. Bulk Save to Database
  log.Printf("[+] Collected %d posts. Ingesting...", len(posts))
  if err := db.BulkIngestRawPosts(context.Background(), conn, posts); err != nil {
    log.Printf("[-] Database ingestion error: %v", err)
    return
  }

  log.Printf("[+] Job Complete. Database connection closed.")
}

func main() {
  _ = godotenv.Load()

  // 1. NATS Setup
  natsURL := os.Getenv("NATS_URL")
  if natsURL == "" {
    natsURL = "nats://127.0.0.1:4222"
  }
  nc, err := nt.NewClient(natsURL)
  if err != nil {
    log.Fatalf("[-] NATS connection error: %v", err)
  }
  defer nc.Close()

  // 2. Initialize Worker with App Service
  app := &App{}
  worker := nt.NewNatsWorker(nc, app)

  // 3. Start Listening for Events
  subject := "scraper.event"
  if err := worker.Start(subject); err != nil {
    log.Fatalf("[-] Failed to start NATS worker: %v", err)
  }

  log.Printf("[+] Scraper Service Online. Listening on subject: %s", subject)

  // Keep process alive
  select {}
}

