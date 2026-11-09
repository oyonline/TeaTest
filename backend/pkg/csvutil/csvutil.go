package csvutil

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// ImportResult 导入结果
type ImportResult struct {
	SuccessCount int      `json:"success_count"`
	FailCount    int      `json:"fail_count"`
	FailReasons  []string `json:"fail_reasons,omitempty"`
}

// QuestionImport 题库导入结构
type QuestionImport struct {
	QuestionNo       int
	BankName         string
	QuestionTypeCode string
	QuestionTypeName string
	QuestionText     string
	OptionA          string
	OptionB          string
	OptionC          string
	OptionD          string
	OptionE          string
	CorrectAnswer    string
}

// ReadCSVWithEncoding 读取 CSV 文件（自动处理 UTF-8 和 GBK 编码）
func ReadCSVWithEncoding(reader io.Reader) ([][]string, error) {
	// 读取所有数据
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 尝试检测和处理 BOM
	data = removeBOM(data)

	// 首先尝试 UTF-8 解码
	records, err := tryDecodeUTF8(data)
	if err == nil {
		return records, nil
	}

	// UTF-8 解码失败，尝试 GBK 解码
	records, err = tryDecodeGBK(data)
	if err == nil {
		return records, nil
	}

	return nil, fmt.Errorf("无法识别文件编码，请确保文件为 UTF-8 或 GBK 编码")
}

// removeBOM 移除 UTF-8 BOM
func removeBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

// tryDecodeUTF8 尝试 UTF-8 解码
func tryDecodeUTF8(data []byte) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1 // 允许可变字段数
	return reader.ReadAll()
}

// tryDecodeGBK 尝试 GBK 解码
func tryDecodeGBK(data []byte) ([][]string, error) {
	// 使用 GB18030 解码器（兼容 GBK）
	decoder := simplifiedchinese.GB18030.NewDecoder()
	transformed, _, err := transform.Bytes(decoder, data)
	if err != nil {
		return nil, fmt.Errorf("GBK 解码失败: %v", err)
	}

	reader := csv.NewReader(bytes.NewReader(transformed))
	reader.FieldsPerRecord = -1
	return reader.ReadAll()
}

// ParseQuestions 解析 CSV 数据。未知题型或任一非法行都会记录失败，调用方必须整批拒绝。
func ParseQuestions(records [][]string, startRow int) ([]QuestionImport, *ImportResult) {
	result := &ImportResult{FailCount: 0, FailReasons: []string{}}
	var questions []QuestionImport

	for i := startRow; i < len(records); i++ {
		row := records[i]
		if len(row) < 11 {
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 字段不足", i+1))
			continue
		}

		for j := range row {
			row[j] = strings.TrimSpace(row[j])
		}
		var questionNo int
		if _, err := fmt.Sscanf(row[0], "%d", &questionNo); err != nil || questionNo <= 0 {
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 题号无效", i+1))
			continue
		}
		if row[4] == "" {
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 题干不能为空", i+1))
			continue
		}
		if row[10] == "" {
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 正确答案不能为空", i+1))
			continue
		}

		questionTypeCode := strings.ToLower(strings.TrimSpace(row[2]))
		correctAnswer := row[10]
		fail := func(reason string) {
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: %s", i+1, reason))
		}

		switch questionTypeCode {
		case "single_choice":
			correctAnswer = strings.ToUpper(correctAnswer)
			if len(correctAnswer) != 1 || !strings.Contains("ABCDE", correctAnswer) {
				fail("单选题正确答案必须是 A、B、C、D 或 E 中的单个字母")
				continue
			}
		case "multiple_choice":
			correctAnswer = strings.ToUpper(correctAnswer)
			if len(correctAnswer) < 2 || len(correctAnswer) > 5 {
				fail("多选题正确答案必须是 2-5 个字母的组合")
				continue
			}
			valid := true
			for _, ch := range correctAnswer {
				if ch < 'A' || ch > 'E' {
					valid = false
					break
				}
			}
			if !valid {
				fail("多选题正确答案只能包含字母 A、B、C、D、E")
				continue
			}
		case "true_false":
			correctAnswer = strings.ToUpper(correctAnswer)
			if correctAnswer != "A" && correctAnswer != "B" {
				fail("判断题正确答案必须是 A 或 B")
				continue
			}
		case "fill_blank", "short_answer":
			correctAnswer = strings.TrimSpace(row[10])
			if correctAnswer == "" {
				fail("文字题参考答案不能为空")
				continue
			}
			if utf8.RuneCountInString(correctAnswer) > 4000 {
				fail("文字题参考答案不能超过4000个字符")
				continue
			}
		default:
			fail("未知题型，仅支持 single_choice/multiple_choice/true_false/fill_blank/short_answer")
			continue
		}

		isChoice := questionTypeCode == "single_choice" || questionTypeCode == "multiple_choice" || questionTypeCode == "true_false"
		if isChoice && (row[5] == "" || row[6] == "") {
			fail("至少要有 A、B 两个选项")
			continue
		}

		typeName := row[3]
		if typeName == "" {
			typeName = defaultQuestionTypeName(questionTypeCode)
		}
		questions = append(questions, QuestionImport{
			QuestionNo:       questionNo,
			BankName:         row[1],
			QuestionTypeCode: questionTypeCode,
			QuestionTypeName: typeName,
			QuestionText:     row[4],
			OptionA:          row[5],
			OptionB:          row[6],
			OptionC:          row[7],
			OptionD:          row[8],
			OptionE:          row[9],
			CorrectAnswer:    correctAnswer,
		})
		result.SuccessCount++
	}

	return questions, result
}

func defaultQuestionTypeName(code string) string {
	switch code {
	case "single_choice":
		return "单选题"
	case "multiple_choice":
		return "多选题"
	case "true_false":
		return "判断题"
	case "fill_blank":
		return "填空题"
	case "short_answer":
		return "简答题"
	default:
		return code
	}
}

// ReadCSVFile 读取 CSV 文件并返回解析后的数据
func ReadCSVFile(file io.Reader) ([][]string, error) {
	return ReadCSVWithEncoding(file)
}

// CreateCSVScanner 创建 CSV 扫描器（用于大文件流式处理）
func CreateCSVScanner(reader io.Reader) *bufio.Scanner {
	return bufio.NewScanner(reader)
}
