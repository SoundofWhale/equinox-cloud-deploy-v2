package ocr

import (
	"os/exec"
	"strings"
	"testing"
)

// A mock AI query function to test summarization
func mockAIQuery(role, prompt string) (string, error) {
	return "Mock summary:\n- Item 1\n- Item 2", nil
}

func TestExtractTextMock(t *testing.T) {
	_, err := exec.LookPath("tesseract")
	if err != nil {
		t.Skip("tesseract not found in PATH, skipping OCR test")
	}

	t.Log("Tesseract is available, but no test image provided")
}

func TestSummarizeWithAI(t *testing.T) {
	extracted := "This is a test document with [NAME] and [EMAIL] anonymized."
	summary, err := SummarizeWithAI(extracted, mockAIQuery)
	if err != nil {
		t.Fatalf("SummarizeWithAI failed: %v", err)
	}
	if !strings.Contains(summary, "Mock summary") {
		t.Errorf("Expected mock summary, got %s", summary)
	}
	t.Logf("✅ Summary generated successfully: %s", summary)
}
