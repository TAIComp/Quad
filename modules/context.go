// File: modules/context.go
package modules

import (
    "github.com/blevesearch/bleve"
    speech "cloud.google.com/go/speech/apiv1"
    // Add other necessary imports here
)

// UserContext holds information about the user's learning context.
type UserContext struct {
    Interests    []string `json:"interests"`
    EnglishLevel string   `json:"english_level"`
}

// AppContext holds shared resources for handlers.
type AppContext struct {
    Index        bleve.Index
    SpeechClient *speech.Client
    KB           KnowledgeBase
    UserManager  *UserManager
}
