package services

import (
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
