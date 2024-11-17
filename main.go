// File: main.go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/joho/godotenv"
    "tutorai-go/modules"

    speech "cloud.google.com/go/speech/apiv1"
    "google.golang.org/api/option"
)

func main() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Log environment variables to ensure they are loaded correctly
    log.Printf("GOOGLE_APPLICATION_CREDENTIALS: %s", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
    log.Printf("GEMINI_API_KEY: %s", os.Getenv("GEMINI_API_KEY"))

    // Initialize UserManager with hardcoded users
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        log.Fatal("JWT_SECRET is not set in .env file")
    }

    userManager, err := modules.NewUserManager(jwtSecret)
    if err != nil {
        log.Fatalf("Error initializing UserManager: %v", err)
    }

    // Load Knowledge Base
    kb, err := modules.LoadKnowledgeBase("knowledge_base.json")
    if err != nil {
        log.Fatalf("Error loading knowledge base: %v", err)
    }

    // Create Bleve Index
    index, err := modules.CreateBleveIndex(kb, "bleve_index")
    if err != nil {
        log.Fatalf("Error creating Bleve index: %v", err)
    }

    // Initialize Google Speech Client with credentials file
    speechClient, err := speech.NewClient(context.Background(), option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
    if err != nil {
        log.Fatalf("Failed to create Speech client: %v", err)
    }
    defer speechClient.Close()

    // Initialize AppContext
    appContext := modules.AppContext{
        Index:        index,
        SpeechClient: speechClient,
        KB:           kb,
        UserManager:  userManager,
    }

    // Setup HTTP Routes and Handlers
    http.HandleFunc("/api/register", appContext.HandleRegister)
    http.HandleFunc("/api/login", appContext.HandleLogin)
    http.Handle("/api/get-response", appContext.AuthMiddleware(http.HandlerFunc(appContext.HandleGetResponse)))
    http.Handle("/api/conversation-history", appContext.AuthMiddleware(http.HandlerFunc(appContext.HandleConversationHistory)))
    http.Handle("/api/me", appContext.AuthMiddleware(http.HandlerFunc(appContext.HandleMe)))
    http.HandleFunc("/api/text-to-speech", modules.TextToSpeechHandler)

    // Add favicon handler
    http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "static/favicon.ico")
    })

    // Serve static files
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/", fs)

    // Start periodic cleanup in a separate goroutine
    go periodicCleanup()

    // Start the HTTP Server
    log.Println("Server started at :1111")
    if err := http.ListenAndServe(":1111", nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func periodicCleanup() {
    ticker := time.NewTicker(15 * time.Minute)  // Adjust interval as needed
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Clean up tmp directory
            files, err := filepath.Glob("tmp/*_audio_*.webm")
            if err != nil {
                log.Printf("Error finding audio files during periodic cleanup: %v", err)
                continue
            }

            // Group files by username
            userFiles := make(map[string][]string)
            for _, file := range files {
                parts := strings.Split(filepath.Base(file), "_")
                if len(parts) >= 2 {
                    username := parts[0]
                    userFiles[username] = append(userFiles[username], file)
                }
            }

            // Keep only most recent file for each user
            for username, files := range userFiles {
                cleanupOldAudioFiles("tmp", username)
            }

            // Remove files older than 24 hours
            threshold := time.Now().Add(-24 * time.Hour)
            for _, file := range files {
                info, err := os.Stat(file)
                if err != nil {
                    continue
                }
                if info.ModTime().Before(threshold) {
                    os.Remove(file)
                }
            }
        }
    }
}

func cleanupOldAudioFiles(directory, username string) {
    files, err := filepath.Glob(filepath.Join(directory, fmt.Sprintf("%s_audio_*.webm", username)))
    if err != nil {
        log.Printf("Error finding old audio files: %v", err)
        return
    }

    // Keep track of the most recent file
    var mostRecent string
    var mostRecentTime int64

    // Find the most recent file
    for _, file := range files {
        // Extract timestamp from filename
        parts := strings.Split(file, "_")
        if len(parts) < 3 {
            continue
        }
        
        // Remove .webm extension and convert to int64
        timestamp := strings.TrimSuffix(parts[len(parts)-1], ".webm")
        t, err := strconv.ParseInt(timestamp, 10, 64)
        if err != nil {
            continue
        }

        if t > mostRecentTime {
            mostRecentTime = t
            mostRecent = file
        }
    }

    // Delete all files except the most recent
    for _, file := range files {
        if file != mostRecent {
            if err := os.Remove(file); err != nil {
                log.Printf("Error removing old audio file %s: %v", file, err)
            }
        }
    }
}
