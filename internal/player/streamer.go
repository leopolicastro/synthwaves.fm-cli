package player

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/vorbis"
	"github.com/gopxl/beep/v2/wav"
)

// audioStream holds a decoded audio stream and metadata about its capabilities.
type audioStream struct {
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	Seekable bool // false for Icecast/infinite streams
}

// readSeekCloser wraps an io.ReadSeeker with a no-op Close.
// Unlike io.NopCloser, this preserves the Seek method.
type readSeekCloser struct {
	io.ReadSeeker
}

func (readSeekCloser) Close() error { return nil }

// openAudioStream fetches audio from a URL and decodes it.
// For finite files (Content-Length present), it buffers the entire response
// to enable seeking. For infinite streams (Icecast), it pipes the body directly.
func openAudioStream(url string) (*audioStream, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching audio: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("audio fetch returned %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	codec := detectCodec(contentType, url)
	if codec == "" {
		resp.Body.Close()
		return nil, fmt.Errorf("unsupported audio format: %s", contentType)
	}

	// Icecast streams are detected by the icy-metaint header or no Content-Length.
	// Everything else (file URLs, ActiveStorage blobs) gets buffered for seek.
	isIcecast := resp.Header.Get("icy-metaint") != "" ||
		(resp.ContentLength < 0 && strings.Contains(strings.ToLower(contentType), "mpeg"))
	seekable := !isIcecast

	var rc io.ReadCloser
	if seekable {
		// Buffer the full response so the decoder can seek.
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("buffering audio: %w", err)
		}
		rc = readSeekCloser{bytes.NewReader(data)}
	} else {
		rc = resp.Body
	}

	streamer, format, err := decode(rc, codec)
	if err != nil {
		rc.Close()
		return nil, fmt.Errorf("decoding %s: %w", codec, err)
	}

	return &audioStream{
		Streamer: streamer,
		Format:   format,
		Seekable: seekable,
	}, nil
}

type codec string

const (
	codecMP3    codec = "mp3"
	codecFLAC   codec = "flac"
	codecVorbis codec = "vorbis"
	codecWAV    codec = "wav"
)

// detectCodec determines the audio codec from Content-Type header or URL extension.
func detectCodec(contentType, url string) codec {
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "mpeg"), strings.Contains(ct, "mp3"):
		return codecMP3
	case strings.Contains(ct, "flac"):
		return codecFLAC
	case strings.Contains(ct, "ogg"), strings.Contains(ct, "vorbis"):
		return codecVorbis
	case strings.Contains(ct, "wav"), strings.Contains(ct, "wave"):
		return codecWAV
	}

	ext := strings.ToLower(path.Ext(url))
	// Strip query params from extension
	if idx := strings.IndexByte(ext, '?'); idx >= 0 {
		ext = ext[:idx]
	}
	switch ext {
	case ".mp3":
		return codecMP3
	case ".flac":
		return codecFLAC
	case ".ogg":
		return codecVorbis
	case ".wav":
		return codecWAV
	}

	// Default to MP3 -- most common for web audio
	return codecMP3
}

func decode(rc io.ReadCloser, c codec) (beep.StreamSeekCloser, beep.Format, error) {
	switch c {
	case codecMP3:
		return mp3.Decode(rc)
	case codecFLAC:
		return flac.Decode(rc)
	case codecVorbis:
		return vorbis.Decode(rc)
	case codecWAV:
		return wav.Decode(rc)
	default:
		return nil, beep.Format{}, fmt.Errorf("unknown codec: %s", c)
	}
}
