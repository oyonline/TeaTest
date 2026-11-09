package csvutil

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestParseQuestionsFiveTypesPreserveText(t *testing.T) {
	records := [][]string{
		{"题号", "题库", "题型编码", "题型名称", "题干", "A", "B", "C", "D", "E", "参考答案"},
		{"1", "通用", "single_choice", "单选题", "单选", "甲", "乙", "", "", "", "a"},
		{"2", "通用", "multiple_choice", "多选题", "多选", "甲", "乙", "丙", "", "", "ca"},
		{"3", "通用", "true_false", "判断题", "判断", "对", "错", "", "", "", "b"},
		{"4", "通用", "fill_blank", "填空题", "填空", "", "", "", "", "", " Tea\n第一行 "},
		{"5", "通用", "short_answer", "简答题", "简答", "", "", "", "", "", "参考答案\nKeep Case"},
	}
	questions, result := ParseQuestions(records, 1)
	if result.FailCount != 0 || result.SuccessCount != 5 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if questions[0].CorrectAnswer != "A" || questions[1].CorrectAnswer != "CA" || questions[2].CorrectAnswer != "B" {
		t.Fatalf("objective answers were not normalized: %#v", questions[:3])
	}
	if questions[3].CorrectAnswer != "Tea\n第一行" || questions[4].CorrectAnswer != "参考答案\nKeep Case" {
		t.Fatalf("text answers changed: %q / %q", questions[3].CorrectAnswer, questions[4].CorrectAnswer)
	}
	if questions[3].OptionA != "" || questions[4].QuestionTypeName != "简答题" {
		t.Fatalf("text question options/name unexpected: %#v", questions[3:])
	}
}

func TestParseQuestionsRejectsUnknownAndOversize(t *testing.T) {
	records := [][]string{
		{"题号", "题库", "题型编码", "题型名称", "题干", "A", "B", "C", "D", "E", "参考答案"},
		{"1", "通用", "ordering", "排序题", "未知", "甲", "乙", "", "", "", "A"},
		{"2", "通用", "short_answer", "简答题", "超长", "", "", "", "", "", strings.Repeat("字", 4001)},
	}
	questions, result := ParseQuestions(records, 1)
	if len(questions) != 0 || result.FailCount != 2 || result.SuccessCount != 0 {
		t.Fatalf("invalid rows must all fail: questions=%d result=%+v", len(questions), result)
	}
}

func TestReadCSVQuotedMultilineText(t *testing.T) {
	data := "题号,题库,题型编码,题型名称,题干,选项A,选项B,选项C,选项D,选项E,参考答案\n" +
		"1,通用,fill_blank,填空题,\"第一行\n第二行\",,,,,,参考答案\n"
	records, err := ReadCSVWithEncoding(strings.NewReader(data))
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	questions, result := ParseQuestions(records, 1)
	if result.FailCount != 0 || len(questions) != 1 {
		t.Fatalf("unexpected parse result: %+v", result)
	}
	if !strings.Contains(questions[0].QuestionText, "\n") {
		t.Fatalf("multiline question text was lost: %q", questions[0].QuestionText)
	}
}

func TestTextAnswerRuneLimitBoundary(t *testing.T) {
	exact := strings.Repeat("😀", 4000)
	if utf8.RuneCountInString(exact) != 4000 {
		t.Fatal("test fixture is not 4000 runes")
	}
	records := [][]string{
		{"题号", "题库", "题型编码", "题型名称", "题干", "A", "B", "C", "D", "E", "参考答案"},
		{"1", "通用", "short_answer", "简答题", "边界", "", "", "", "", "", exact},
	}
	_, result := ParseQuestions(records, 1)
	if result.FailCount != 0 {
		t.Fatalf("4000-rune answer should be accepted: %+v", result)
	}
	records[1][10] = exact + "😀"
	_, result = ParseQuestions(records, 1)
	if result.FailCount != 1 {
		t.Fatalf("4001-rune answer should be rejected: %+v", result)
	}
}
