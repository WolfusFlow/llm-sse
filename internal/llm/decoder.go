package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type streamLine struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type StreamingDecoder struct {
	reader *bufio.Reader
}

func NewStreamingDecoder(r io.Reader) *StreamingDecoder {
	return &StreamingDecoder{reader: bufio.NewReader(r)}
}

func (d *StreamingDecoder) NextChunk() (string, error) {
	for {
		line, err := d.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		raw := strings.TrimPrefix(line, "data: ")
		if raw == "[DONE]" {
			return "", io.EOF
		}

		var parsed streamLine
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return "", fmt.Errorf("unmarshal stream chunk: %w", err)
		}
		if len(parsed.Choices) == 0 {
			continue
		}

		return parsed.Choices[0].Delta.Content, nil
	}
}
