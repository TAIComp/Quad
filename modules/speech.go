// File: modules/speech.go
package modules

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
    "strconv"

    "github.com/blevesearch/bleve"
    "github.com/blevesearch/bleve/mapping"
    "github.com/google/generative-ai-go/genai"
    speech "cloud.google.com/go/speech/apiv1"
    texttospeech "cloud.google.com/go/texttospeech/apiv1"
    speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
    texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
    "google.golang.org/api/option"
)

// KnowledgeBase is a map where keys are topic names and values are Topic structs.
type KnowledgeBase map[string]Topic

// Topic represents a single topic in the knowledge base.
type Topic struct {
    Text   string   `json:"text"`
    Images []string `json:"images"`
}

// TopicResult struct for search results.
type TopicResult struct {
    TopicName string
    Text      string
    Images    []string
}

// Message represents a single exchange between the user and the AI.
type Message struct {
    Role    string `json:"role"`    // "user" or "assistant"
    Content string `json:"content"` // The message content
}

// ConversationHistory holds the list of messages exchanged.
type ConversationHistory struct {
    Messages []Message `json:"messages"`
}

// AddMessage adds a new message to the conversation history and limits the history size.
func (ch *ConversationHistory) AddMessage(role, content string) {
    ch.Messages = append(ch.Messages, Message{Role: role, Content: content})
    // Limit history to the last 20 messages (10 user and 10 assistant)
    if len(ch.Messages) > 20 {
        ch.Messages = ch.Messages[len(ch.Messages)-20:]
    }
}

// GetPrompt constructs the prompt for the AI by concatenating the history.
func (ch *ConversationHistory) GetPrompt(modelDescription string, userContext UserContext) string {
    var promptBuilder strings.Builder
    promptBuilder.WriteString(modelDescription)
    promptBuilder.WriteString("\n\n")
    promptBuilder.WriteString(fmt.Sprintf("User's interests: %v\n", userContext.Interests))
    promptBuilder.WriteString(fmt.Sprintf("User's English level: %s\n", userContext.EnglishLevel))
    promptBuilder.WriteString("\n")
    for _, msg := range ch.Messages {
        if msg.Role == "user" {
            promptBuilder.WriteString("User: " + msg.Content + "\n")
        } else if msg.Role == "assistant" {
            promptBuilder.WriteString("AI: " + msg.Content + "\n")
        }
    }
    promptBuilder.WriteString("AI:") // Prompt the AI for the next response
    return promptBuilder.String()
}

// LoadConversationHistory loads the conversation history from a JSON file.
func LoadConversationHistory(filename string) (*ConversationHistory, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        if os.IsNotExist(err) {
            // If the file doesn't exist, return an empty history
            return &ConversationHistory{Messages: []Message{}}, nil
        }
        return nil, err
    }
    var history ConversationHistory
    if err := json.Unmarshal(data, &history); err != nil {
        return nil, err
    }
    return &history, nil
}

// SaveToFile saves the conversation history to a JSON file.
func (ch *ConversationHistory) SaveToFile(filename string) error {
    data, err := json.Marshal(ch)
    if err != nil {
        return err
    }
    return os.WriteFile(filename, data, 0644)
}

// LoadKnowledgeBase loads the knowledge base from a JSON file.
func LoadKnowledgeBase(filename string) (KnowledgeBase, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read knowledge base file: %v", err)
    }

    var kb KnowledgeBase
    if err := json.Unmarshal(data, &kb); err != nil {
        return nil, fmt.Errorf("failed to unmarshal knowledge base JSON: %v", err)
    }

    return kb, nil
}

// CreateBleveIndex creates or opens a Bleve index from the knowledge base.
func CreateBleveIndex(kb KnowledgeBase, indexPath string) (bleve.Index, error) {
    // Define the mapping.
    indexMapping := buildIndexMapping()

    // Attempt to create a new index.
    index, err := bleve.New(indexPath, indexMapping)
    if err != nil {
        // If the index already exists, open it.
        if err == bleve.ErrorIndexPathExists {
            index, err = bleve.Open(indexPath)
            if err != nil {
                return nil, fmt.Errorf("failed to open existing index: %v", err)
            }
        } else {
            return nil, fmt.Errorf("failed to create index: %v", err)
        }
    }

    // Index each topic.
    for topicName, topic := range kb {
        doc := map[string]interface{}{
            "topic":  topicName,
            "text":   topic.Text,
            "images": topic.Images,
        }
        if err := index.Index(topicName, doc); err != nil {
            return nil, fmt.Errorf("failed to index topic %s: %v", topicName, err)
        }
    }

    return index, nil
}

// buildIndexMapping defines the Bleve index mapping with English language analyzer.
func buildIndexMapping() mapping.IndexMapping {
    // Define field mappings.
    textFieldMapping := bleve.NewTextFieldMapping()
    textFieldMapping.Analyzer = "en"
    textFieldMapping.Store = true

    imagesFieldMapping := bleve.NewTextFieldMapping()
    imagesFieldMapping.Analyzer = "en"
    imagesFieldMapping.Store = true

    topicFieldMapping := bleve.NewTextFieldMapping()
    topicFieldMapping.Analyzer = "en"
    topicFieldMapping.Store = true

    // Define document mapping.
    docMapping := bleve.NewDocumentMapping()
    docMapping.AddFieldMappingsAt("topic", topicFieldMapping)
    docMapping.AddFieldMappingsAt("text", textFieldMapping)
    docMapping.AddFieldMappingsAt("images", imagesFieldMapping)

    // Define index mapping.
    indexMapping := bleve.NewIndexMapping()
    indexMapping.DefaultMapping = docMapping
    indexMapping.DefaultAnalyzer = "en"

    return indexMapping
}

// SearchKnowledgeBase performs a search on the Bleve index.
func SearchKnowledgeBase(index bleve.Index, query string, interests []string) ([]TopicResult, error) {
    // Combine the query with interests.
    combinedQuery := fmt.Sprintf("%s %s", query, strings.Join(interests, " "))

    // Create a search query.
    searchQuery := bleve.NewMatchQuery(combinedQuery)
    searchQuery.SetFuzziness(1) // Allow slight misspellings.

    // Create a search request.
    search := bleve.NewSearchRequest(searchQuery)
    search.Size = 5 // Return top 5 results.

    // Execute the search.
    searchResult, err := index.Search(search)
    if err != nil {
        return nil, fmt.Errorf("failed to execute search: %v", err)
    }

    // Collect the results.
    var results []TopicResult
    for _, hit := range searchResult.Hits {
        var topicResult TopicResult
        topicResult.TopicName = hit.ID

        doc, err := index.Document(hit.ID)
        if err != nil {
            return nil, fmt.Errorf("failed to retrieve document ID %s: %v", hit.ID, err)
        }

        for _, field := range doc.Fields {
            switch field.Name() {
            case "text":
                topicResult.Text = string(field.Value())
            case "images":
                var images []string
                if err := json.Unmarshal(field.Value(), &images); err != nil {
                    log.Printf("Failed to unmarshal images for topic %s: %v", hit.ID, err)
                    images = []string{}
                }
                topicResult.Images = images
            }
        }
        results = append(results, topicResult)
    }

    return results, nil
}

// GetResponseFromAudioFile processes an audio file and returns the AI's response, user input text, and audio data.
func GetResponseFromAudioFile(ctx context.Context, index bleve.Index, history *ConversationHistory, userContext UserContext, audioFile string) (string, string, string, error) {
    // Convert speech to text.
    inputText, err := speechToText(ctx, audioFile)
    if err != nil {
        return "", "", "", fmt.Errorf("failed to convert speech to text: %v", err)
    }
    inputText = strings.TrimSpace(inputText)
    log.Printf("Transcribed Input: %s", inputText)

    // Normalize input for consistent matching.
    normalizedInput := strings.ToLower(inputText)

    // Check for termination phrases.
    if strings.Contains(normalizedInput, "stop conversation") || strings.Contains(normalizedInput, "finish conversation") {
        log.Println("Conversation ended by user request.")
        history.AddMessage("assistant", "Conversation ended as per your request.")
        return "Conversation ended as per your request.", inputText, "", nil
    }

    // Add user input to history.
    history.AddMessage("user", inputText)

    // Preprocess the input text (optional).
    processedText := preprocessText(inputText)

    // Search the knowledge base using Bleve.
    topics, err := SearchKnowledgeBase(index, processedText, userContext.Interests)
    if err != nil {
        return "", inputText, "", fmt.Errorf("failed to search knowledge base: %v", err)
    }

    var aiResponse string

    // Modify the response selection logic
    if len(topics) > 0 && isRelevantTopic(topics[0], inputText) {
        // Use the text from the first matching topic only if it's relevant
        aiResponse = topics[0].Text
    } else {
        // If no relevant topic is found, generate response with Gemini
        geminiResponse, err := generateContentWithGemini(ctx, history, userContext)
        if err != nil {
            return "", inputText, "", fmt.Errorf("failed to generate content with Gemini: %v", err)
        }
        aiResponse = extractResponseFromGemini(geminiResponse)
        if aiResponse == "" {
            return "", inputText, "", fmt.Errorf("no valid response generated")
        }
    }

    log.Printf("AI Response: %s", aiResponse)

    // Add AI response to history.
    history.AddMessage("assistant", aiResponse)

    // Generate audio for AI's response.
    audioData, err := GenerateAudio(ctx, aiResponse)
    if err != nil {
        return "", inputText, "", fmt.Errorf("failed to generate audio: %v", err)
    }

    return aiResponse, inputText, audioData, nil
}

// preprocessText performs basic NLP preprocessing on the input text.
func preprocessText(text string) string {
    // Convert to lowercase.
    text = strings.ToLower(text)
    // Remove punctuation (optional).
    text = removePunctuation(text)
    // Additional preprocessing steps can be added here.
    return text
}

// removePunctuation removes punctuation from a string.
func removePunctuation(text string) string {
    var sb strings.Builder
    for _, r := range text {
        if !strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", r) {
            sb.WriteRune(r)
        }
    }
    return sb.String()
}

// speechToText converts the audio file to text using Google Speech-to-Text API.
func speechToText(ctx context.Context, filename string) (string, error) {
    // Initialize Google Cloud Speech client with credentials
    client, err := speech.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
    if err != nil {
        return "", fmt.Errorf("failed to create Speech client: %v", err)
    }
    defer client.Close()

    data, err := os.ReadFile(filename)
    if err != nil {
        return "", fmt.Errorf("failed to read audio file: %v", err)
    }

    req := &speechpb.RecognizeRequest{
        Config: &speechpb.RecognitionConfig{
            Encoding:        speechpb.RecognitionConfig_WEBM_OPUS,
            SampleRateHertz: 48000,
            LanguageCode:    "en-US",
        },
        Audio: &speechpb.RecognitionAudio{
            AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
        },
    }

    resp, err := client.Recognize(ctx, req)
    if err != nil {
        return "", fmt.Errorf("failed to recognize speech: %v", err)
    }

    var transcription strings.Builder
    for _, result := range resp.Results {
        for _, alt := range result.Alternatives {
            transcription.WriteString(alt.Transcript)
        }
    }
    return transcription.String(), nil
}

// GenerateAudio converts text to speech and returns Base64-encoded audio data.
func GenerateAudio(ctx context.Context, text string) (string, error) {
    // Initialize Google Cloud Text-to-Speech client with credentials
    client, err := texttospeech.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
    if err != nil {
        return "", fmt.Errorf("failed to create Text-to-Speech client: %v", err)
    }
    defer client.Close()

    // Prepare the text input for speech synthesis.
    req := &texttospeechpb.SynthesizeSpeechRequest{
        Input: &texttospeechpb.SynthesisInput{
            InputSource: &texttospeechpb.SynthesisInput_Text{
                Text: text,
            },
        },
        Voice: &texttospeechpb.VoiceSelectionParams{
            LanguageCode: "en-US",
            Name:         "en-US-Casual-K", // Set the specific voice name.
        },
        AudioConfig: &texttospeechpb.AudioConfig{
            AudioEncoding: texttospeechpb.AudioEncoding_MP3,
        },
    }

    // Perform the text-to-speech request.
    resp, err := client.SynthesizeSpeech(ctx, req)
    if err != nil {
        return "", fmt.Errorf("failed to synthesize speech: %v", err)
    }

    // Encode the audio content to Base64.
    audioBase64 := base64.StdEncoding.EncodeToString(resp.AudioContent)

    return audioBase64, nil
}

// generateContentWithGemini generates content using the Gemini API based on the conversation history.
func generateContentWithGemini(ctx context.Context, history *ConversationHistory, userContext UserContext) (*genai.GenerateContentResponse, error) {
    // Access your API key from the environment variable.
    geminiAPIKey := os.Getenv("GEMINI_API_KEY")
    if geminiAPIKey == "" {
        return nil, fmt.Errorf("GEMINI_API_KEY not set in environment variables")
    }

    client, err := genai.NewClient(ctx, option.WithAPIKey(geminiAPIKey))
    if err != nil {
        return nil, fmt.Errorf("failed to create Gemini client: %v", err)
    }

    // Define the model description with improved context handling
    modelDescription := fmt.Sprintf(`
    You are Quad, an AI-powered online teacher dedicated to making learning fun and engaging.
    Your focus areas include: %v.
    You communicate in English tailored to a %s-level learner.

    Role and Personality:
    - Be patient, encouraging, and supportive.
    - Use interactive teaching methods to engage the learner.

    Guidelines:
    1. **Accuracy & Relevance:** Always provide correct and pertinent information about the requested topic.
    2. **Admit Uncertainty:** If unsure about a topic, acknowledge it and offer general insights or suggest reliable sources.
    3. **Focused Responses:** Address the user's specific questions directly and relevantly.
    4. **Use Examples:** Incorporate examples, analogies, or visual descriptions to clarify concepts when appropriate.
    5. **Clarity & Conciseness:** Keep explanations clear and to the point, avoiding unnecessary complexity.
    6. **Adapt to Complexity:** For simple queries, respond in 1-2 sentences. For complex topics, provide detailed explanations while maintaining clarity.
    7. **Encourage Engagement:** Ask follow-up questions or prompt the user to think about the topic to enhance understanding.
    8. **Feedback:** Occasionally check if the user understands or needs further clarification.

    Conversation Context:
    The user is interested in learning about various topics. Provide accurate and relevant information about the specific topic they inquire about. Keep responses brief unless the topic's complexity necessitates a more detailed explanation.

    Additional Instructions:
    - Maintain a friendly and approachable tone.
    - Tailor explanations to match the user's English proficiency level.
    `, userContext.Interests, userContext.EnglishLevel)
    
    // Construct the prompt with conversation history and user context.
    prompt := history.GetPrompt(modelDescription, userContext)

    // Generate content using Gemini API.
    model := client.GenerativeModel("gemini-1.5-flash")
    resp, err := model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        return nil, fmt.Errorf("failed to generate content with Gemini: %v", err)
    }

    return resp, nil
}

// extractResponseFromGemini parses the Gemini API response to extract the generated text.
func extractResponseFromGemini(resp *genai.GenerateContentResponse) string {
    if resp != nil && len(resp.Candidates) > 0 {
        // Assuming the first candidate is the best response.
        if resp.Candidates[0].Content != nil {
            // Create a variable to store the combined text.
            var result strings.Builder

            // Iterate over the parts of the content.
            for _, part := range resp.Candidates[0].Content.Parts {
                if txt, ok := part.(genai.Text); ok {
                    result.WriteString(string(txt)) // Append the text to the result.
                }
            }

            // Return the concatenated text.
            return result.String()
        }
    }
    return ""
}

// TextToSpeechHandler handles the text-to-speech conversion and returns the audio data.
func TextToSpeechHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse the request body
    var request struct {
        Text string `json:"text"`
    }
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Generate audio using the existing GenerateAudio function
    ctx := context.Background()
    audioBase64, err := GenerateAudio(ctx, request.Text)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to generate audio: %v", err), http.StatusInternalServerError)
        return
    }

    // Decode base64 to binary
    audioData, err := base64.StdEncoding.DecodeString(audioBase64)
    if err != nil {
        http.Error(w, "Failed to decode audio data", http.StatusInternalServerError)
        return
    }

    // Set response headers
    w.Header().Set("Content-Type", "audio/mp3")
    w.Header().Set("Content-Length", strconv.Itoa(len(audioData)))

    // Write the audio data to the response
    if _, err := w.Write(audioData); err != nil {
        log.Printf("Error writing audio response: %v", err)
    }
}

// Add new helper function to check topic relevance
func isRelevantTopic(topic TopicResult, query string) bool {
    // Convert both to lowercase for case-insensitive comparison
    topicLower := strings.ToLower(topic.TopicName)
    queryLower := strings.ToLower(query)

    // Calculate similarity score (simple word overlap for now)
    queryWords := strings.Fields(queryLower)
    topicWords := strings.Fields(topicLower)

    matchCount := 0
    for _, queryWord := range queryWords {
        for _, topicWord := range topicWords {
            if queryWord == topicWord {
                matchCount++
            }
        }
    }

    // Require at least 30% word overlap for relevance
    threshold := 0.3
    similarity := float64(matchCount) / float64(len(queryWords))
    return similarity >= threshold
}
