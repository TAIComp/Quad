// package version

// import (
//     "context"
//     "fmt"
//     "log"
//     "os"
//     "os/exec"
//     "strings"
//     "time"

//     "cloud.google.com/go/speech/apiv1"
//     "github.com/google/generative-ai-go/genai"
//     "github.com/joho/godotenv"
//     speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
//     "google.golang.org/api/option"
// )

// // SpeechToResponse records audio, converts it to text, and gets a response from Gemini API.
// func SpeechToResponse() (string, error) {
//     // Load environment variables from .env file
//     if err := godotenv.Load(); err != nil {
//         return "", fmt.Errorf("error loading .env file: %v", err)
//     }

//     ctx := context.Background()

//     // Set path to your service account credentials
//     credentialsPath := "/home/manu/Desktop/TutorAI Go/tutorai-438115-991792d2aeb8.json" // Consider making this configurable

//     // Initialize Speech-to-Text client
//     speechClient, err := speech.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
//     if err != nil {
//         return "", fmt.Errorf("failed to create Speech client: %v", err)
//     }
//     defer speechClient.Close()

//     // Record audio input
//     fmt.Println("Recording... Please speak into your microphone.")
//     audioFile := "input.raw"
//     if err := recordAudio(audioFile); err != nil {
//         return "", fmt.Errorf("failed to record audio: %v", err)
//     }
//     fmt.Println("Recording complete.")

//     // Convert speech to text
//     inputText, err := speechToText(ctx, speechClient, audioFile)
//     if err != nil {
//         return "", fmt.Errorf("failed to convert speech to text: %v", err)
//     }
//     fmt.Printf("You said: %s\n", inputText)

//     // Send input text to Gemini API and get response
//     geminiResponse, err := generateContentWithGemini(ctx, inputText)
//     if err != nil {
//         return "", fmt.Errorf("failed to generate content with Gemini: %v", err)
//     }

//     // Extract response text
//     responseText := extractGeminiResponse(geminiResponse)
//     fmt.Printf("Gemini Response: %s\n", responseText)

//     return responseText, nil
// }

// // recordAudio records audio from the microphone and saves it to the specified file.
// func recordAudio(filename string) error {
//     duration := 5 // seconds
//     ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
//     defer cancel()

//     cmd := exec.CommandContext(ctx, "parec", "--format", "s16le", "--rate", "44100", "--channels", "1")
//     outFile, err := os.Create(filename)
//     if err != nil {
//         return fmt.Errorf("failed to create audio file: %v", err)
//     }
//     defer outFile.Close()

//     cmd.Stdout = outFile
//     cmd.Stderr = os.Stderr

//     if err := cmd.Run(); err != nil && ctx.Err() != context.DeadlineExceeded {
//         return fmt.Errorf("failed to record audio: %v", err)
//     }
//     return nil
// }

// // speechToText converts the audio file to text using Google Speech-to-Text API.
// func speechToText(ctx context.Context, client *speech.Client, filename string) (string, error) {
//     data, err := os.ReadFile(filename)
//     if err != nil {
//         return "", fmt.Errorf("failed to read audio file: %v", err)
//     }

//     req := &speechpb.RecognizeRequest{
//         Config: &speechpb.RecognitionConfig{
//             Encoding:        speechpb.RecognitionConfig_LINEAR16,
//             SampleRateHertz: 44100,
//             LanguageCode:    "en-US",
//         },
//         Audio: &speechpb.RecognitionAudio{
//             AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
//         },
//     }

//     resp, err := client.Recognize(ctx, req)
//     if err != nil {
//         return "", fmt.Errorf("failed to recognize speech: %v", err)
//     }

//     var transcription strings.Builder
//     for _, result := range resp.Results {
//         for _, alt := range result.Alternatives {
//             transcription.WriteString(alt.Transcript)
//         }
//     }
//     return transcription.String(), nil
// }

// // generateContentWithGemini generates content using the Gemini API based on the input text.
// func generateContentWithGemini(ctx context.Context, inputText string) (*genai.GenerateContentResponse, error) {
//     // Access your API key from the environment variable
//     client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
//     if err != nil {
//         return nil, err
//     }
//     defer client.Close()

//     // Generate content using Gemini API
//     model := client.GenerativeModel("gemini-1.5-flash")
//     resp, err := model.GenerateContent(ctx, genai.Text(inputText))
//     if err != nil {
//         return nil, err
//     }

//     return resp, nil
// }

// // extractGeminiResponse extracts the response text from Gemini API response.
// func extractGeminiResponse(resp *genai.GenerateContentResponse) string {
//     if resp != nil && len(resp.Candidates) > 0 {
//         // Assuming the candidates contain a field that holds the generated text
//         return resp.Candidates[0].Content // Change 'Content' to the actual field name if different
//     }
//     return "No response received from Gemini."
// }
