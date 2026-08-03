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
		{Name: "张三", Password: strings.Repeat("密", 101)},
	}
	for _, input := range tests {
		if err := validateExamUserInput(&input); err == nil {
			t.Fatalf("invalid input should return an error: %#v", input)
		}
	}
}
