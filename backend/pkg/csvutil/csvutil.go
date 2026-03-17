package csvutil

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

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

// ParseQuestions 解析 CSV 数据为题库记录
func ParseQuestions(records [][]string, startRow int) ([]QuestionImport, *ImportResult) {
	result := &ImportResult{
		SuccessCount: 0,
		FailCount:    0,
		FailReasons:  []string{},
	}

	var questions []QuestionImport

	// 跳过标题行，从 startRow 开始解析
	for i := startRow; i < len(records); i++ {
		row := records[i]
		if len(row) < 11 {
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 字段不足", i+1))
			continue
		}

		// 清理字段
		for j := range row {
			row[j] = strings.TrimSpace(row[j])
		}

		// 解析题号
		var questionNo int
		_, err := fmt.Sscanf(row[0], "%d", &questionNo)
		if err != nil || questionNo <= 0 {
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 题号无效", i+1))
			continue
		}

		// 验证必填字段
		if row[4] == "" { // 题干
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 题干不能为空", i+1))
			continue
		}

		if row[10] == "" { // 正确答案
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 正确答案不能为空", i+1))
			continue
		}

		// 获取题型编码和正确答案（统一转小写处理）
		questionTypeCode := strings.TrimSpace(strings.ToLower(row[2]))
		correctAnswer := strings.ToUpper(row[10])

		// 严格按照三种题型编码判断，并验证答案格式
		var finalTypeCode string
		switch questionTypeCode {
		case "multiple_choice":
			finalTypeCode = "multiple_choice"
			// 多选题验证：允许多个字母组合（2-5个，如 AB、ACD）
			if len(correctAnswer) < 2 || len(correctAnswer) > 5 {
				result.FailCount++
				result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 多选题正确答案必须是 2-5 个字母的组合", i+1))
				continue
			}
			// 验证每个字符都是 A-E
			valid := true
			for _, ch := range correctAnswer {
				if ch != 'A' && ch != 'B' && ch != 'C' && ch != 'D' && ch != 'E' {
					valid = false
					break
				}
			}
			if !valid {
				result.FailCount++
				result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 多选题正确答案只能包含字母 A、B、C、D、E", i+1))
				continue
			}
		case "true_false":
			finalTypeCode = "true_false"
			// 判断题验证：只能有 A 或 B
			if correctAnswer != "A" && correctAnswer != "B" {
				result.FailCount++
				result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 判断题正确答案必须是 A 或 B", i+1))
				continue
			}
		default:
			// 单选题（包括 single_choice 和任何其他值）
			finalTypeCode = "single_choice"
			// 单选题验证：只能有一个正确答案
			if len(correctAnswer) != 1 || (correctAnswer != "A" && correctAnswer != "B" && correctAnswer != "C" && correctAnswer != "D" && correctAnswer != "E") {
				result.FailCount++
				result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 单选题正确答案必须是 A、B、C、D 或 E 中的单个字母", i+1))
				continue
			}
		}

		// 至少要有 A、B 两个选项
		if row[5] == "" || row[6] == "" { // OptionA, OptionB
			result.FailCount++
			result.FailReasons = append(result.FailReasons, fmt.Sprintf("第%d行: 至少要有 A、B 两个选项", i+1))
			continue
		}

		q := QuestionImport{
			QuestionNo:       questionNo,
			BankName:         row[1],
			QuestionTypeCode: finalTypeCode,
			QuestionTypeName: row[3],
			QuestionText:     row[4],
			OptionA:          row[5],
			OptionB:          row[6],
			OptionC:          row[7],
			OptionD:          row[8],
			OptionE:          row[9],
			CorrectAnswer:    correctAnswer,
		}

		questions = append(questions, q)
		result.SuccessCount++
	}

	return questions, result
}

// ReadCSVFile 读取 CSV 文件并返回解析后的数据
func ReadCSVFile(file io.Reader) ([][]string, error) {
	return ReadCSVWithEncoding(file)
}

// CreateCSVScanner 创建 CSV 扫描器（用于大文件流式处理）
func CreateCSVScanner(reader io.Reader) *bufio.Scanner {
	return bufio.NewScanner(reader)
}
