// package version

// import (
//     "context"
//     "fmt"
//     "io/ioutil"
//     "os"
//     "os/exec"

//     "github.com/joho/godotenv"
//     texttospeech "cloud.google.com/go/texttospeech/apiv1"
//     texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
// )

// // ResponseToVoice converts the response text to speech and plays the audio.
// func ResponseToVoice(textInput string) error {
//     // Load environment variables from .env file
//     if err := godotenv.Load(); err != nil {
//         return fmt.Errorf("error loading .env file: %v", err)
//     }

//     // Set Google Application Credentials
//     credentialsPath := "./tutorai-438115-991792d2aeb8.json" // Ensure this path is correct
//     if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsPath); err != nil {
//         return fmt.Errorf("failed to set GOOGLE_APPLICATION_CREDENTIALS: %v", err)
//     }

//     // Initialize Google Cloud Text-to-Speech client
//     ctx := context.Background()
//     client, err := texttospeech.NewClient(ctx)
//     if err != nil {
//         return fmt.Errorf("failed to create Text-to-Speech client: %v", err)
//     }
//     defer client.Close()

//     // Prepare the text input for speech synthesis
//     req := &texttospeechpb.SynthesizeSpeechRequest{
//         Input: &texttospeechpb.SynthesisInput{
//             InputSource: &texttospeechpb.SynthesisInput_Text{
//                 Text: textInput,
//             },
//         },
//         // Select the voice
//         Voice: &texttospeechpb.VoiceSelectionParams{
//             LanguageCode: "en-US",
//             Name:         "en-US-Casual-K", // Use the Casual_K voice
//         },
//         // Select the audio file format
//         AudioConfig: &texttospeechpb.AudioConfig{
//             AudioEncoding: texttospeechpb.AudioEncoding_MP3,
//         },
//     }

//     // Perform the text-to-speech request
//     resp, err := client.SynthesizeSpeech(ctx, req)
//     if err != nil {
//         return fmt.Errorf("failed to synthesize speech: %v", err)
//     }

//     // Save the audio to a file
//     outputFile := "output.mp3"
//     if err := ioutil.WriteFile(outputFile, resp.AudioContent, 0644); err != nil {
//         return fmt.Errorf("failed to write the audio file: %v", err)
//     }

//     fmt.Printf("Audio content written to file: %s\n", outputFile)

//     // Play the audio file using mpg123
//     if err := playAudio(outputFile); err != nil {
//         return fmt.Errorf("failed to play audio: %v", err)
//     }

//     return nil
// }

// // playAudio plays an audio file using system's audio player (mpg123 for Linux).
// func playAudio(filename string) error {
//     cmd := "mpg123 " + filename
//     if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
//         return fmt.Errorf("failed to play audio: %v", err)
//     }
//     return nil
// }
