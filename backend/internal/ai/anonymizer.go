package ai

import "regexp"

// anonymizerPattern defines a regex-to-placeholder mapping.
type anonymizerPattern struct {
	regex   *regexp.Regexp
	replace string
}

// patterns covers Russian + English PII: names, emails, phones, money, locations.
var patterns = []anonymizerPattern{
	// Emails (match first — most specific)
	{regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`), "[EMAIL]"},

	// Phone numbers — RU: +7/8, international: +XX
	{regexp.MustCompile(`(\+7|8)[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}`), "[PHONE]"},
	{regexp.MustCompile(`\+\d{1,3}[\s\-]?\(?\d{2,4}\)?[\s\-]?\d{3,4}[\s\-]?\d{2,4}`), "[PHONE]"},

	// Monetary amounts (руб, $, €, ₽, тыс, млн)
	{regexp.MustCompile(`\d[\d\s,.]*\s*(руб|рублей|USD|EUR|₽|\$|€|тыс|млн|thousand|million)`), "[AMOUNT]"},

	// Locations (city/street patterns in RU and EN)
	{regexp.MustCompile(`(ул\.|пр\.|г\.|пер\.|city|street|avenue|ave\.?|rd\.?)\s+[А-ЯA-Z][а-яa-z\s]+`), "[LOCATION]"},

	// Names after Russian prepositions/particles (с Иваном, для Анны, от Петра)
	// Using explicit whitespace/start-of-line instead of \b which fails with Cyrillic
	{regexp.MustCompile(`(?:^|\s)(для|от|с|у|к)\s+[А-ЯЁ][а-яё]+`), " [NAME]"},
	// English honorifics
	{regexp.MustCompile(`\b(Mr|Mrs|Ms|Dr|Prof)\s+[A-Z][a-z]+`), "[NAME]"},
}

// AnonymizeText replaces PII patterns with placeholders.
// The original text must NEVER be logged or stored after this call.
func AnonymizeText(input string) string {
	result := input
	for _, p := range patterns {
		result = p.regex.ReplaceAllString(result, p.replace)
	}
	return result
}
