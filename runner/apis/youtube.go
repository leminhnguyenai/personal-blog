package apis

import (
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

func GetYtbData(id string) (Snippet, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	url := fmt.Sprintf(
		"https://youtube.googleapis.com/youtube/v3/videos?part=snippet&id=%s&key=%s",
		id,
		apiKey,
	)

	res, err := http.Get(url)
	if err != nil {
		return Snippet{}, err
	}

	if res.StatusCode != http.StatusOK {
		return Snippet{}, fmt.Errorf("Bad request, code:%d\n", res.StatusCode)
	}

	var ytbResponse YtbResponse

	if err = json.NewDecoder(res.Body).Decode(&ytbResponse); err != nil {
		return Snippet{}, err
	}

	if len(ytbResponse.Items) == 0 {
		return Snippet{}, fmt.Errorf("Error retrieving video\n")
	}

	ytbVidData := ytbResponse.Items[0].Snippet

	return ytbVidData, nil
}
