package theme

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// RenderImageITerm2 fetches an image URL and renders it inline using
// the iTerm2 image protocol. Works in iTerm2, WezTerm, and Kitty.
// Returns empty string if the terminal doesn't support it or fetch fails.
// width is in terminal columns.
func RenderImageITerm2(url string, width int) string {
	if url == "" {
		return ""
	}

	// Check if we're in a terminal that supports inline images
	term := os.Getenv("TERM_PROGRAM")
	if term != "iTerm.app" && term != "WezTerm" {
		// Fall back to just showing the URL
		return ""
	}

	// Fetch the image
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	// iTerm2 inline image protocol:
	// ESC ] 1337 ; File=[args] : base64data BEL
	var b strings.Builder
	b.WriteString("\033]1337;File=inline=1")
	b.WriteString(fmt.Sprintf(";width=%d", width))
	b.WriteString(";preserveAspectRatio=1")
	b.WriteString(":")
	b.WriteString(encoded)
	b.WriteString("\a")

	return b.String()
}
