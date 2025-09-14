package main

import (
	"chat_example/util"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/genai"
)

func main() {
	// [START chat_streaming_with_images]
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	chat, err := client.Chats.Create(ctx, "gemini-2.0-flash", nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	for chunk, err := range chat.SendMessageStream(ctx, genai.Part{
		Text: "Hello, I'm interested in learning about musical instruments. Can I show you one?"}) {
		if err != nil {
			log.Fatal(err)
		}
		log.Println(chunk.Text())
	}

	log.Println(strings.Repeat("_", 64))

	image, err := client.Files.UploadFromPath(
		ctx,
		filepath.Join(util.GetMedia(), "saxophone.png"),
		&genai.UploadFileConfig{
			MIMEType: "image/png",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	parts := make([]genai.Part, 2)
	parts[0] = genai.Part{Text: "What family of instruments does this instrument belong to?"}
	parts[1] = genai.Part{
		FileData: &genai.FileData{
			FileURI:  image.URI,
			MIMEType: image.MIMEType,
		},
	}

	for chunk, err := range chat.SendMessageStream(ctx, parts...) {
		if err != nil {
			log.Fatal(err)
		}
		log.Println(chunk.Text())
	}
	log.Println(strings.Repeat("_", 64))

}
