package gcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

const maxWorkers = 10

type GoogleTTSClient struct {
	client   *texttospeech.Client
	speaking bool
}

func (c *GoogleTTSClient) Close() {
	c.client.Close()
}

func NewGoogleTTSClient(ctx context.Context) (*GoogleTTSClient, error) {
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS client: %v", err)
	}
	return &GoogleTTSClient{client: client, speaking: false}, nil
}

// GetSpeech converts text to speech in order
func (c *GoogleTTSClient) GetSpeech(text string) ([]byte, error) {
	fmt.Println("Will be speaking:", text)
	ctx := context.Background()
	startTime := time.Now()

	sentences := splitIntoSentences(text)
	numSentences := len(sentences)

	// Use min(numSentences, maxWorkers) to avoid extra goroutines
	workerCount := min(numSentences, maxWorkers)

	// Ensure ordered output
	results := make([][]byte, numSentences)
	jobs := make(chan int, numSentences)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for index := range jobs {
				start := time.Now()
				audioData := processSentence(ctx, c.client, sentences[index])
				if audioData != nil {
					results[index] = audioData // Store at correct index
				}
				fmt.Printf("Worker %d processed sentence %d in: %v\n", workerID, index, time.Since(start))
			}
		}(i)
	}

	// Assign jobs
	for i := 0; i < numSentences; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()

	// Merge audio properly with silence between sentences
	finalAudio := mergeAudioFiles(results)

	fmt.Printf("Total processing time: %v\n", time.Since(startTime))

	return finalAudio, nil
}

func (c *GoogleTTSClient) Speaking(speaking bool) {
	c.speaking = speaking
}

// Splits a paragraph into sentences, also handling Hindi punctuation
func splitIntoSentences(text string) []string {
	sentences := strings.Split(text, ". ")
	for i, sentence := range sentences {
		sentences[i] = strings.TrimSpace(sentence)
	}
	return sentences
}

// Processes a sentence using Google Cloud TTS
func processSentence(ctx context.Context, client *texttospeech.Client, sentence string) []byte {
	input := &texttospeechpb.SynthesisInput{
		InputSource: &texttospeechpb.SynthesisInput_Text{Text: sentence},
	}

	voice := &texttospeechpb.VoiceSelectionParams{
		LanguageCode: "hi-IN",
		// Name:         "hi-IN-Standard-D",
		Name: "hi-IN-Chirp3-HD-Aoede",
	}

	audioConfig := &texttospeechpb.AudioConfig{
		AudioEncoding:   texttospeechpb.AudioEncoding_LINEAR16,
		SampleRateHertz: 8000,
	}

	resp, err := client.SynthesizeSpeech(ctx, &texttospeechpb.SynthesizeSpeechRequest{
		Input:       input,
		Voice:       voice,
		AudioConfig: audioConfig,
	})
	if err != nil {
		log.Printf("Error synthesizing speech: %v", err)
		return nil
	}

	return resp.AudioContent[44:] //44: removes wav header that causes mouse click sound at beging
}

// Merges audio files in correct order with silence between sentences
func mergeAudioFiles(results [][]byte) []byte {
	var finalAudio []byte

	// 0.5 seconds of silence buffer (adjust as needed)
	// silenceBuffer := make([]byte, 8000/2)

	for _, audioData := range results {
		if audioData != nil {
			finalAudio = append(finalAudio, audioData...)
			finalAudio = append(finalAudio, generateSilence(0.5, 8000)...) // Add silence between sentences
		}
	}

	return finalAudio
}

// Returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateSilence(durationSeconds float64, sampleRate int) []byte {
	numSamples := int(float64(sampleRate) * durationSeconds)
	silence := make([]byte, numSamples*2) // 2 bytes per sample for 16-bit PCM
	return silence
}
