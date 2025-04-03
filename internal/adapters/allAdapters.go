package adapters

type AllAdapter struct{
	GroqAdapter GroqAdapter
	YoutubeAdapter YouTubeAdapter
}

func NewAllAdapter(groqApiKey string, YoutubeApiKey string) AllAdapter{
	return AllAdapter{
		GroqAdapter: NewGroqAdapter(groqApiKey),
		YoutubeAdapter: NewYouTubeAdapter(YoutubeApiKey),
	}
}