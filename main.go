package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	token := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(token)
	ctx := context.Background()

	concepts, err := getTenComputerScienceConcepts(ctx, client)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return	
	}

	fmt.Println("The 10 most important Computer Science Concepts are: ", concepts)

	createImagesDirirectory()

	generateAndSaveImages(ctx, client, concepts)

	fmt.Println("Images generated and saved in the images directory")
}

func getTenComputerScienceConcepts(ctx context.Context, client *openai.Client) ([]string, error) {
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "As I Software Developer, I want to know the 10 most important Computer Science Concepts. I want just the concept title like 'Data structures' or 'Networks', for example. Please provide the answer in the following format:\n\n1 - Concept 1\n2 - Concept 2\n...\n10 - Concept 10",
				},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("ChatCompletion error: %w", err)
	}

	concepts := strings.Split(resp.Choices[0].Message.Content, "\n")
	for i := range concepts {
		concepts[i] = strings.TrimSpace(concepts[i])
		concepts[i] = strings.TrimPrefix(concepts[i], fmt.Sprintf("%d - ", i+1))
	}

	return concepts, nil
}

func createImagesDirirectory() {
	if _, err := os.Stat("images"); err == nil {
		os.RemoveAll("images")
		os.Mkdir("images", 0755)
	}
}

func generateAndSaveImages(ctx context.Context, client *openai.Client, concepts []string) error {
	for _, concept := range concepts {
		fmt.Printf("Generating image for concept %s\n", concept)

		reqUrl := openai.ImageRequest{
			Prompt:         fmt.Sprintf("Thumbnail for the Computer Science class about %s", concept),
			Size:           openai.CreateImageSize1024x1024,
			ResponseFormat: openai.CreateImageResponseFormatURL,
			N:              1,
		}

		respUrl, err := client.CreateImage(ctx, reqUrl)
		if err != nil {
			return fmt.Errorf("Image creation error: %w", err)
		}

		if err := saveImage(ctx, respUrl.Data[0].URL, fmt.Sprintf("images/%s.png", concept)); err != nil {
			return fmt.Errorf("Image save error: %w", err)
		}
	}

	return nil
}

func saveImage(ctx context.Context, url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Image download error: %w", err)
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("File creation error: %w", err)
	}

	if _, err := file.ReadFrom(resp.Body); err != nil {
		return fmt.Errorf("File write error: %w", err)
	}

	return nil
}