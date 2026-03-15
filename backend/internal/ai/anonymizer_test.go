package ai

import "testing"

func TestAnonymizeEmail(t *testing.T) {
	input := "Send a report to ivan.petrov@company.com by 5pm"
	result := AnonymizeText(input)
	if result == input {
		t.Fatal("email was not anonymized")
	}
	if !containsPlaceholder(result, "[EMAIL]") {
		t.Fatalf("expected [EMAIL] placeholder, got %q", result)
	}
	t.Logf("✅ Email anonymized: %s", result)
}

func TestAnonymizePhoneRU(t *testing.T) {
	input := "Позвони по +7 (999) 123-45-67 после обеда"
	result := AnonymizeText(input)
	if !containsPlaceholder(result, "[PHONE]") {
		t.Fatalf("expected [PHONE] placeholder, got %q", result)
	}
	t.Logf("✅ Phone anonymized: %s", result)
}

func TestAnonymizeMoney(t *testing.T) {
	input := "Бюджет проекта 500 000 рублей на разработку"
	result := AnonymizeText(input)
	if !containsPlaceholder(result, "[AMOUNT]") {
		t.Fatalf("expected [AMOUNT] placeholder, got %q", result)
	}
	t.Logf("✅ Amount anonymized: %s", result)
}

func TestAnonymizeLocation(t *testing.T) {
	input := "Встреча в г. Москва на ул. Тверская"
	result := AnonymizeText(input)
	if !containsPlaceholder(result, "[LOCATION]") {
		t.Fatalf("expected [LOCATION] placeholder, got %q", result)
	}
	t.Logf("✅ Location anonymized: %s", result)
}

func TestAnonymizeName(t *testing.T) {
	input := "Отправить отчёт для Анны до конца дня"
	result := AnonymizeText(input)
	if !containsPlaceholder(result, "[NAME]") {
		t.Fatalf("expected [NAME] placeholder, got %q", result)
	}
	t.Logf("✅ Name anonymized: %s", result)
}

func TestAnonymizeMultiple(t *testing.T) {
	input := "Mr Smith at john@acme.com from city London owes 5000 USD. Call +7 999 111-22-33"
	result := AnonymizeText(input)
	for _, tag := range []string{"[EMAIL]", "[PHONE]", "[AMOUNT]", "[NAME]"} {
		if !containsPlaceholder(result, tag) {
			t.Errorf("expected %s placeholder in %q", tag, result)
		}
	}
	t.Logf("✅ Multi-pattern anonymized: %s", result)
}

func TestAnonymizeCleanText(t *testing.T) {
	input := "Review the architecture document and fix bugs"
	result := AnonymizeText(input)
	if result != input {
		t.Fatalf("clean text was modified: got %q", result)
	}
	t.Log("✅ Clean text unchanged")
}

func containsPlaceholder(text, placeholder string) bool {
	return len(text) > 0 && text != "" && indexOf(text, placeholder) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
