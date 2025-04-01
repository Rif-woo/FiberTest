# scripts/get_transcript.py

import sys
from youtube_transcript_api import YouTubeTranscriptApi

def get_transcript(video_id):
    try:
        # Tente d'abord le français, puis l'anglais
        transcript = YouTubeTranscriptApi.get_transcript(video_id, languages=['fr', 'en'])
        text = "\n".join([entry["text"] for entry in transcript])
        return text
    except Exception as e:
        return f"ERROR: {str(e)}"

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("ERROR: No video ID provided.")
    else:
        video_id = sys.argv[1]
        print(get_transcript(video_id))
