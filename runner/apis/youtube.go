package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Snippet struct {
	PublishedAt time.Time `json:"publishedAt"`
	ChannelID   string    `json:"channelId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Thumbnails  struct {
		Default struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"default"`
		Medium struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"medium"`
		High struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"high"`
		Standard struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"standard"`
	} `json:"thumbnails"`
	ChannelTitle string `json:"channelTitle"`
}

type YtbResponse struct {
	Items []struct {
		ID      string  `json:"id"`
		Snippet Snippet `json:"snippet"`
	} `json:"items"`
}

func GetYtbData(ctx context.Context, id string) (Snippet, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	url := fmt.Sprintf(
		"https://youtube.googleapis.com/youtube/v3/videos?part=snippet&id=%s&key=%s",
		id,
		apiKey,
	)

	type Response struct {
		snippet Snippet
		error   error
	}

	respChan := make(chan Response)

	go func() {
		res, err := http.Get(url)
		if err != nil {
			respChan <- Response{error: err}
		}

		if res.StatusCode != http.StatusOK {
			respChan <- Response{error: fmt.Errorf("Bad request, code:%d\n", res.StatusCode)}
		}

		var ytbResponse YtbResponse

		if err = json.NewDecoder(res.Body).Decode(&ytbResponse); err != nil {
			respChan <- Response{error: err}
		}

		if len(ytbResponse.Items) == 0 {
			respChan <- Response{error: fmt.Errorf("Error retrieving video\n")}
		}

		ytbVidData := ytbResponse.Items[0].Snippet

		respChan <- Response{ytbVidData, nil}
	}()

	for {
		select {
		case <-ctx.Done():
			return Snippet{}, fmt.Errorf("Unexpected error occured\n")
		case res := <-respChan:
			return res.snippet, res.error
		}
	}
}
