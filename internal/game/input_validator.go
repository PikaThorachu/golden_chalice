package game

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// InputValidator handles all input validation and sanitization
type InputValidator struct {
	maxInputLength          int
	allowedCommands         map[string]bool
	chineseDirectionPattern *regexp.Regexp
	englishDirectionPattern *regexp.Regexp
}

// ValidationResult contains the result of input validation
type ValidationResult struct {
	IsValid         bool
	Sanitized       string
	ErrorMsg        string
	ErrorMsgPinyin  string
	ErrorMsgEnglish string
}

// NewInputValidator creates a new input validator with default settings
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxInputLength: 200,
		allowedCommands: map[string]bool{
			// Chinese commands
			"帮助": true, "背包": true, "状态": true, "查看": true, "看": true,
			"退出": true, "保存": true, "加载": true, "存档列表": true,
			"装备": true, "卸下": true, "使用": true, "丢弃": true, "拿": true, "取": true,
			// English commands
			"help": true, "inventory": true, "i": true, "status": true,
			"look": true, "quit": true, "exit": true, "save": true,
			"load": true, "saves": true, "equip": true, "unequip": true,
			"use": true, "drop": true, "take": true,
		},
		chineseDirectionPattern: regexp.MustCompile(`^往[北南东西出][走去]$`),
		englishDirectionPattern: regexp.MustCompile(`^(go|walk)\s+(north|south|east|west|out)$`),
	}
}

// ValidateAndSanitize performs comprehensive input validation
func (iv *InputValidator) ValidateAndSanitize(input string) ValidationResult {
	// Check for empty input
	if input == "" {
		return ValidationResult{
			IsValid:         false,
			Sanitized:       "",
			ErrorMsg:        "输入不能为空",
			ErrorMsgPinyin:  "Shūrù bùnéng wéi kōng",
			ErrorMsgEnglish: "Input cannot be empty",
		}
	}

	// Check maximum length
	if utf8.RuneCountInString(input) > iv.maxInputLength {
		return ValidationResult{
			IsValid:         false,
			Sanitized:       "",
			ErrorMsg:        "输入太长，请保持在200个字符以内",
			ErrorMsgPinyin:  "Shūrù tài cháng, qǐng bǎochí zài 200 gè zìfú yǐnèi",
			ErrorMsgEnglish: "Input too long, please keep within 200 characters",
		}
	}

	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Check for suspicious characters (optional - for security)
	if iv.containsSuspiciousCharacters(sanitized) {
		return ValidationResult{
			IsValid:         false,
			Sanitized:       "",
			ErrorMsg:        "输入包含无效字符",
			ErrorMsgPinyin:  "Shūrù bāohán wúxiào zìfú",
			ErrorMsgEnglish: "Input contains invalid characters",
		}
	}

	return ValidationResult{
		IsValid:   true,
		Sanitized: sanitized,
		ErrorMsg:  "",
	}
}

// ValidateMovementCommand validates movement commands specifically
func (iv *InputValidator) ValidateMovementCommand(input string) ValidationResult {
	sanitized := strings.TrimSpace(input)

	// Check Chinese format
	if strings.HasPrefix(sanitized, "往") {
		if iv.chineseDirectionPattern.MatchString(sanitized) {
			return ValidationResult{
				IsValid:   true,
				Sanitized: sanitized,
			}
		}

		// Provide specific error for malformed Chinese movement
		return ValidationResult{
			IsValid:         false,
			Sanitized:       sanitized,
			ErrorMsg:        "格式错误。请使用: 往北走, 往南走, 往东走, 往西走, 往出走",
			ErrorMsgPinyin:  "Géshì cuòwù. Qǐng shǐyòng: wǎng běi zǒu, wǎng nán zǒu, wǎng dōng zǒu, wǎng xī zǒu, wǎng chū zǒu",
			ErrorMsgEnglish: "Invalid format. Please use: go north, go south, go east, go west, go out",
		}
	}

	// Check English format
	lowerInput := strings.ToLower(sanitized)
	if strings.HasPrefix(lowerInput, "go ") || strings.HasPrefix(lowerInput, "walk ") {
		if iv.englishDirectionPattern.MatchString(lowerInput) {
			return ValidationResult{
				IsValid:   true,
				Sanitized: sanitized,
			}
		}

		return ValidationResult{
			IsValid:         false,
			Sanitized:       sanitized,
			ErrorMsg:        "Invalid direction. Please use: go north, go south, go east, go west, go out",
			ErrorMsgPinyin:  "Wúxiào fāngxiàng. Qǐng shǐyòng: go north, go south, go east, go west, go out",
			ErrorMsgEnglish: "Invalid direction. Please use: go north, go south, go east, go west, go out",
		}
	}

	return ValidationResult{
		IsValid:         false,
		Sanitized:       sanitized,
		ErrorMsg:        "未知命令",
		ErrorMsgPinyin:  "Wèizhī mìnglìng",
		ErrorMsgEnglish: "Unknown command",
	}
}

// ValidateItemCommand validates take/use/drop/equip commands
func (iv *InputValidator) ValidateItemCommand(input string) (command string, itemName string, isValid bool, errorMsg string) {
	sanitized := strings.TrimSpace(input)

	// Define command prefixes
	commandPrefixes := map[string]string{
		"拿": "take", "取": "take", "take": "take",
		"使用": "use", "use": "use",
		"装备": "equip", "equip": "equip",
		"丢弃": "drop", "drop": "drop",
	}

	for prefix, cmd := range commandPrefixes {
		if strings.HasPrefix(sanitized, prefix) {
			// Extract item name
			itemName = strings.TrimPrefix(sanitized, prefix)
			itemName = strings.TrimSpace(itemName)

			// Validate item name is not empty
			if itemName == "" {
				return "", "", false, "请指定物品名称"
			}

			// Validate item name length
			if utf8.RuneCountInString(itemName) > 50 {
				return "", "", false, "物品名称太长"
			}

			return cmd, itemName, true, ""
		}
	}

	return "", "", false, "无效的物品命令"
}

// ValidateSaveCommand validates save/load commands
func (iv *InputValidator) ValidateSaveCommand(input string) (command string, slotName string, isValid bool, errorMsg string) {
	sanitized := strings.ToLower(strings.TrimSpace(input))

	// Check for save command
	if strings.HasPrefix(sanitized, "save ") || strings.HasPrefix(sanitized, "保存 ") {
		parts := strings.Fields(sanitized)
		if len(parts) >= 2 {
			slotName = parts[1]
			// Validate slot name
			if !iv.isValidSlotName(slotName) {
				return "", "", false, "存档名称只能包含字母、数字和下划线"
			}
			return "save", slotName, true, ""
		}
		return "save", "autosave", true, "" // Default slot
	}

	// Check for load command
	if strings.HasPrefix(sanitized, "load ") || strings.HasPrefix(sanitized, "加载 ") {
		parts := strings.Fields(sanitized)
		if len(parts) >= 2 {
			slotName = parts[1]
			if !iv.isValidSlotName(slotName) {
				return "", "", false, "存档名称只能包含字母、数字和下划线"
			}
			return "load", slotName, true, ""
		}
		return "", "", false, "请指定要加载的存档名称"
	}

	// Check for delete command
	if strings.HasPrefix(sanitized, "delete ") || strings.HasPrefix(sanitized, "删除 ") {
		parts := strings.Fields(sanitized)
		if len(parts) >= 2 {
			slotName = parts[1]
			if !iv.isValidSlotName(slotName) {
				return "", "", false, "存档名称只能包含字母、数字和下划线"
			}
			return "delete", slotName, true, ""
		}
		return "", "", false, "请指定要删除的存档名称"
	}

	return "", "", false, "无效的存档命令"
}

// ValidateNumberInput validates numeric input (for future use)
func (iv *InputValidator) ValidateNumberInput(input string, min, max int) (int, bool, string) {
	sanitized := strings.TrimSpace(input)

	// Check if empty
	if sanitized == "" {
		return 0, false, "输入不能为空"
	}

	// Parse integer
	var num int
	_, err := fmt.Sscanf(sanitized, "%d", &num)
	if err != nil {
		return 0, false, "请输入有效的数字"
	}

	// Check range
	if num < min || num > max {
		return 0, false, fmt.Sprintf("数字必须在 %d 到 %d 之间", min, max)
	}

	return num, true, ""
}

// containsSuspiciousCharacters checks for potentially harmful input
func (iv *InputValidator) containsSuspiciousCharacters(input string) bool {
	// Allow Chinese characters, English letters, numbers, spaces, and common punctuation
	// Block control characters and special escape sequences
	for _, r := range input {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return true
		}
	}
	return false
}

// isValidSlotName validates save slot names
func (iv *InputValidator) isValidSlotName(name string) bool {
	if name == "" {
		return false
	}
	if utf8.RuneCountInString(name) > 30 {
		return false
	}
	// Allow letters, numbers, underscores, and Chinese characters
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_\u4e00-\u9fa5]+$`, name)
	return matched
}

// IsSimpleCommand checks if input is a simple command (no arguments)
func (iv *InputValidator) IsSimpleCommand(input string) bool {
	sanitized := strings.TrimSpace(strings.ToLower(input))
	return iv.allowedCommands[sanitized]
}

// GetCommandCategory returns the category of a command
func (iv *InputValidator) GetCommandCategory(input string) string {
	sanitized := strings.TrimSpace(strings.ToLower(input))

	switch sanitized {
	case "help", "帮助":
		return "help"
	case "inventory", "i", "背包":
		return "inventory"
	case "status", "状态":
		return "status"
	case "look", "看", "查看":
		return "look"
	case "quit", "exit", "退出":
		return "quit"
	case "saves", "存档列表":
		return "list_saves"
	default:
		// Check for movement
		if strings.HasPrefix(sanitized, "往") || strings.HasPrefix(sanitized, "go ") || strings.HasPrefix(sanitized, "walk ") {
			return "movement"
		}
		// Check for item commands
		for prefix := range map[string]bool{"拿": true, "取": true, "take": true, "使用": true, "use": true, "装备": true, "equip": true, "丢弃": true, "drop": true} {
			if strings.HasPrefix(sanitized, prefix) {
				return "item"
			}
		}
		// Check for save commands
		if strings.HasPrefix(sanitized, "save ") || strings.HasPrefix(sanitized, "保存 ") ||
			strings.HasPrefix(sanitized, "load ") || strings.HasPrefix(sanitized, "加载 ") ||
			strings.HasPrefix(sanitized, "delete ") || strings.HasPrefix(sanitized, "删除 ") {
			return "save"
		}
		return "unknown"
	}
}

// SuggestCorrections provides suggestions for mistyped commands
func (iv *InputValidator) SuggestCorrections(input string) []string {
	input = strings.ToLower(strings.TrimSpace(input))
	suggestions := []string{}

	// Common typos and their corrections
	typos := map[string][]string{
		"helo": {"help"}, "halp": {"help"}, "hlep": {"help"},
		"invetory": {"inventory", "i"}, "invent": {"inventory"}, "inv": {"inventory"},
		"stauts": {"status"}, "stat": {"status"},
		"lok": {"look"}, "loook": {"look"},
		"quti": {"quit"}, "qiut": {"quit"},
		"noth": {"go north"}, "soth": {"go south"}, "east": {"go east"}, "wst": {"go west"},
	}

	if corrections, exists := typos[input]; exists {
		suggestions = append(suggestions, corrections...)
	}

	return suggestions
}
