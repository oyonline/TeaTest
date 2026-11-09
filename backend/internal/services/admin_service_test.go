package services

import (
	"strings"
	"testing"
)

func TestValidateExamUserInput(t *testing.T) {
	valid := ExamUserInput{Name: " 张三 ", Password: " 123456 "}
	if err := validateExamUserInput(&valid); err != nil {
		t.Fatalf("valid input returned error: %v", err)
	}
	if valid.Name != "张三" || valid.Password != "123456" {
		t.Fatalf("input was not trimmed: %#v", valid)
	}

	tests := []ExamUserInput{
		{Name: "", Password: "123456"},
		{Name: "张三", Password: ""},
		{Name: strings.Repeat("答", 51), Password: "123456"},
		{Name: "张三", Password: strings.Repeat("密", 25)},
	}
	for _, input := range tests {
		if err := validateExamUserInput(&input); err == nil {
			t.Fatalf("invalid input should return an error: %#v", input)
		}
	}
}

func TestValidateResetExamUserPasswordInput(t *testing.T) {
	tests := []struct {
		name     string
		password string
		confirm  string
		wantErr  bool
	}{
		{"valid", "new-password@@", "new-password@@", false},
		{"preserve spaces", " password ", " password ", false},
		{"empty", "", "", true},
		{"whitespace", " \t ", " \t ", true},
		{"mismatch", "new-password", "other-password", true},
		{"missing confirmation", "new-password", "", true},
		{"72 bytes", strings.Repeat("a", 72), strings.Repeat("a", 72), false},
		{"73 bytes", strings.Repeat("a", 73), strings.Repeat("a", 73), true},
		{"72 unicode bytes", strings.Repeat("密", 24), strings.Repeat("密", 24), false},
		{"75 unicode bytes", strings.Repeat("密", 25), strings.Repeat("密", 25), true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateResetExamUserPasswordInput(ResetExamUserPasswordInput{Password: test.password, ConfirmPassword: test.confirm})
			if (err != nil) != test.wantErr {
				t.Fatalf("validation error = %v, want error = %v", err, test.wantErr)
			}
		})
	}
}
