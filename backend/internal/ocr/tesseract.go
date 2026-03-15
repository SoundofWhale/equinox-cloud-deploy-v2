package ocr

import (
	"fmt"
	"os/exec"
	"strings"
)

// ExtractText runs Tesseract OCR on an image file.
// Supports Russian + English (rus+eng).
func ExtractText(filePath string) (string, error) {
	// Use command-line tesseract for maximum compatibility (no CGO dependency).
	// Output goes to stdout via "-" argument.
	cmd := exec.Command("tesseract", filePath, "stdout", "-l", "rus+eng", "--psm", "3")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tesseract failed on %s: %w", filePath, err)
	}

	text := strings.TrimSpace(string(output))
	if text == "" {
		return "", fmt.Errorf("no text extracted from %s", filePath)
	}

	return text, nil
}

// SummarizeWithAI takes OCR-extracted text, anonymizes it, and generates a summary.
// This is the complete OCR→Anonymize→AI pipeline.
func SummarizeWithAI(extractedText string, queryFunc func(role, prompt string) (string, error)) (string, error) {
	prompt := fmt.Sprintf("Summarize the following document text in 2-3 concise bullet points:\n\n%s", extractedText)

	summary, err := queryFunc("cto", prompt)
	if err != nil {
		return "", fmt.Errorf("AI summarization failed: %w", err)
	}

	return summary, nil
}
