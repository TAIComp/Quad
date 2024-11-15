// File: main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"

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

    // Serve static files
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/", fs)

    // Start the HTTP Server
    log.Println("Server started at :1111")
    if err := http.ListenAndServe(":1111", nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
