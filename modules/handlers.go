// File: modules/handlers.go
package modules

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

// HandleConversationHistory retrieves the user's conversation history
func (app *AppContext) HandleConversationHistory(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        sendErrorResponse(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    username, ok := GetUsernameFromContext(r.Context())
    if !ok {
        sendErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
        return
    }

    log.Printf("Handling conversation history for user: %s", username)

    // Load conversation history
    historyFile := filepath.Join("conversations", fmt.Sprintf("%s.json", username))
    history, err := LoadConversationHistory(historyFile)
    if err != nil {
        log.Printf("Failed to load conversation history for user %s: %v", username, err)
        sendErrorResponse(w, "Failed to load conversation history", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(history)
}

// HandleGetResponse processes the audio and returns AI's response along with audio
func (app *AppContext) HandleGetResponse(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        sendErrorResponse(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    username, ok := GetUsernameFromContext(r.Context())
    if !ok {
        sendErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
        return
    }

    log.Printf("Received /api/get-response request from user: %s", username)

    // Read audio data from the request body
    err := r.ParseMultipartForm(32 << 20) // limit upload size to 32MB
    if err != nil {
        log.Printf("Error parsing multipart form for user %s: %v", username, err)
        sendErrorResponse(w, "Failed to parse form data", http.StatusBadRequest)
        return
    }

    file, _, err := r.FormFile("audio")
    if err != nil {
        log.Printf("Error retrieving audio file for user %s: %v", username, err)
        sendErrorResponse(w, "Failed to retrieve audio file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Ensure the tmp directory exists
    os.MkdirAll("tmp", os.ModePerm)

    // Save the audio file to a temporary location
    audioFilePath := fmt.Sprintf("tmp/%s_audio_%d.webm", username, time.Now().Unix())
    out, err := os.Create(audioFilePath)
    if err != nil {
        log.Printf("Error creating audio file for user %s: %v", username, err)
        sendErrorResponse(w, "Failed to save audio file", http.StatusInternalServerError)
        return
    }
    defer out.Close()

    _, err = io.Copy(out, file)
    if err != nil {
        log.Printf("Error saving audio file for user %s: %v", username, err)
        sendErrorResponse(w, "Failed to save audio file", http.StatusInternalServerError)
        return
    }

    log.Printf("Audio file saved for user %s at %s", username, audioFilePath)

    // Load conversation history
    historyFile := filepath.Join("conversations", fmt.Sprintf("%s.json", username))
    history, err := LoadConversationHistory(historyFile)
    if err != nil {
        log.Printf("Failed to load conversation history for user %s: %v", username, err)
        sendErrorResponse(w, "Failed to load conversation history", http.StatusInternalServerError)
        return
    }

    // Load user context from UserManager (interests and English level)
    user, err := app.UserManager.GetUserContext(username)
    if err != nil {
        log.Printf("Failed to load user context for user %s: %v", username, err)
        sendErrorResponse(w, "Failed to load user context", http.StatusInternalServerError)
        return
    }

    // Create UserContext using 'user'
    userContext := UserContext{
        Interests:    user.Interests,
        EnglishLevel: user.EnglishLevel,
    }

    // Process the audio file and get AI response
    aiResponse, userInputText, audioData, err := GetResponseFromAudioFile(
        r.Context(),
        app.Index,
        history,
        userContext,
        audioFilePath,
    )
    if err != nil {
        log.Printf("Failed to get AI response for user %s: %v", username, err)
        sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
        return
    }

    log.Printf("AI Response for user %s: %s", username, aiResponse)

    // Save updated conversation history
    err = history.SaveToFile(historyFile)
    if err != nil {
        log.Printf("Failed to save conversation history for user %s: %v", username, err)
        sendErrorResponse(w, "Failed to save conversation history", http.StatusInternalServerError)
        return
    }

    log.Printf("Conversation history updated for user %s", username)

    // Send response to client
    response := map[string]string{
        "aiResponse":  aiResponse,
        "userInput":   userInputText,
        "audioBase64": audioData,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// HandleRegister handles user registration (Not necessary with hardcoded users)
func (app *AppContext) HandleRegister(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        sendErrorResponse(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Username     string   `json:"username"`
        Password     string   `json:"password"`
        Interests    []string `json:"interests"`
        EnglishLevel string   `json:"englishLevel"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("Error decoding registration request: %v", err)
        sendErrorResponse(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    if req.Username == "" || req.Password == "" || req.EnglishLevel == "" {
        sendErrorResponse(w, "Username, password, and English level are required", http.StatusBadRequest)
        return
    }

    err := app.UserManager.RegisterUser(req.Username, req.Password, req.Interests, req.EnglishLevel)
    if err != nil {
        log.Printf("Error registering user %s: %v", req.Username, err)
        sendErrorResponse(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Automatically log in the user after registration
    token, err := app.UserManager.AuthenticateUser(req.Username, req.Password)
    if err != nil {
        log.Printf("Error generating token for user %s: %v", req.Username, err)
        sendErrorResponse(w, "Registration successful, but failed to generate token", http.StatusInternalServerError)
        return
    }

    // Set the token as a cookie (Set Secure to false for development)
    http.SetCookie(w, &http.Cookie{
        Name:     "token",
        Value:    token,
        Expires:  time.Now().Add(24 * time.Hour),
        HttpOnly: true,      // Prevent JavaScript access
        Secure:   false,     // Set to false for development; true in production
        Path:     "/",
    })

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Registration successful",
    })
}

// HandleLogin handles user login
func (app *AppContext) HandleLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        sendErrorResponse(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("Error decoding login request: %v", err)
        sendErrorResponse(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    if req.Username == "" || req.Password == "" {
        sendErrorResponse(w, "Username and password are required", http.StatusBadRequest)
        return
    }

    token, err := app.UserManager.AuthenticateUser(req.Username, req.Password)
    if err != nil {
        log.Printf("Authentication failed for user %s: %v", req.Username, err)
        sendErrorResponse(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Set the token as a cookie (Set Secure to false for development)
    http.SetCookie(w, &http.Cookie{
        Name:     "token",
        Value:    token,
        Expires:  time.Now().Add(24 * time.Hour),
        HttpOnly: true,      // Prevent JavaScript access
        Secure:   false,     // Set to false for development; true in production
        Path:     "/",
    })

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Login successful",
    })
}

// HandleMe returns the current authenticated user's information
func (app *AppContext) HandleMe(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        sendErrorResponse(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    username, ok := GetUsernameFromContext(r.Context())
    if !ok {
        sendErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
        return
    }

    // Fetch user information
    user, err := app.UserManager.GetUserContext(username)
    if err != nil {
        sendErrorResponse(w, "Failed to retrieve user info", http.StatusInternalServerError)
        return
    }

    // Prepare response
    response := map[string]interface{}{
        "username":         user.Username,
        "interests":        user.Interests,
        "englishLevel":     user.EnglishLevel,
        "registrationDate": user.RegistrationDate,
        // Add other fields if necessary
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
