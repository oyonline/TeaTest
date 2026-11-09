package services

import (
	"strings"
	"testing"

	"tea-exam/internal/models"
)

func TestNormalizeAnswer(t *testing.T) {
	tests := map[string]string{
		"BAC":   "ABC",
		"b,a,c": "ABC",
		"AABC":  "ABC",
		"无效答案":  "",
	}
	for input, expected := range tests {
		if actual := normalizeAnswer(input); actual != expected {
			t.Fatalf("normalizeAnswer(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestValidateQuestionBankInput(t *testing.T) {
	valid := QuestionBankTypeInput{
		CategoryCode: models.QuestionBankCategoryCompetition,
		Name:         " 茶叶感官审评 ",
		Status:       1,
	}
	if err := validateQuestionBankInput(&valid); err != nil {
		t.Fatalf("valid input returned error: %v", err)
	}
	if valid.Name != "茶叶感官审评" {
		t.Fatalf("name was not trimmed: %q", valid.Name)
	}

	invalidCategory := valid
	invalidCategory.CategoryCode = "custom"
	if err := validateQuestionBankInput(&invalidCategory); err == nil {
		t.Fatal("invalid category should return an error")
	}

	invalidStatus := valid
	invalidStatus.Status = 2
	if err := validateQuestionBankInput(&invalidStatus); err == nil {
		t.Fatal("invalid status should return an error")
	}
}

func TestParseEntryAndEntryCategories(t *testing.T) {
	for _, input := range []string{"", "tea", "TEA", " practice "} {
		if _, err := ParseEntry(input); err != nil {
			t.Fatalf("ParseEntry(%q) returned error: %v", input, err)
		}
	}
	if _, err := ParseEntry("practice2"); err == nil {
		t.Fatal("unknown entry should be rejected")
	}
	if !EntryAllowsCategory(EntryPractice, models.QuestionBankCategoryGeneral) {
		t.Fatal("practice entry should allow general category")
	}
	if EntryAllowsCategory(EntryPractice, models.QuestionBankCategoryCompetition) {
		t.Fatal("practice entry must not allow competition category")
	}
	if EntryAllowsCategory(EntryTea, models.QuestionBankCategoryGeneral) {
		t.Fatal("tea entry must not allow general category")
	}
}

func TestTextAnswerNormalizationAndLength(t *testing.T) {
	got := NormalizeTextAnswer("  第一行\r\n第二行\r第三行  ")
	want := "第一行\n第二行\n第三行"
	if got != want {
		t.Fatalf("NormalizeTextAnswer() = %q, want %q", got, want)
	}
	if _, err := ValidateTextAnswer(" \r\n\t "); err == nil {
		t.Fatal("whitespace-only text answer should be rejected")
	}
	exact := strings.Repeat("😀", MaxTextAnswerRunes)
	if _, err := ValidateTextAnswer(exact); err != nil {
		t.Fatalf("4000 runes should be accepted: %v", err)
	}
	if _, err := ValidateTextAnswer(exact + "😀"); err != ErrTextAnswerTooLong {
		t.Fatalf("4001 runes should return ErrTextAnswerTooLong, got %v", err)
	}

	if graded, correct := GradeTextAnswer(models.QuestionTypeFillBlank, "Tea\nA", "Tea\nA"); !graded || !correct {
		t.Fatal("fill_blank should auto grade exact normalized text")
	}
	if graded, correct := GradeTextAnswer(models.QuestionTypeFillBlank, "Tea", "tea"); !graded || correct {
		t.Fatal("fill_blank must keep case sensitivity")
	}
	if graded, correct := GradeTextAnswer(models.QuestionTypeShortAnswer, "reference", "anything"); graded || correct {
		t.Fatal("short_answer must never auto grade")
	}
}
