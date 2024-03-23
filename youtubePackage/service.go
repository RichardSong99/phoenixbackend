package youtubePackage

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YouTubeService struct {
	Service *youtube.Service
}

func NewYouTubeService(ctx context.Context) (*YouTubeService, error) {

	apiKey := os.Getenv("YOUTUBE_API_KEY")

	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("Error creating YouTube service: %v", err)
	}

	return &YouTubeService{
		Service: service,
	}, nil
}
