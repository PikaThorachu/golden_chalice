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

func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxInputLength: 200,
		allowedCommands: map[string]bool{
			// Chinese commands
			"帮助": true, "背包": true, "状态": true, "查看": true, "看": true,
			"退出": true, "退": true, // Add "退" here
			"保存": true, "加载": true, "存档列表": true,
			"装备": true, "卸下": true, "使用": true, "丢弃": true, "拿": true, "取": true,
			// English commands
			"help": true, "inventory": true, "i": true, "status": true,
			"look": true, "quit": true, "exit": true, "save": true,
			"load": true, "saves": true, "equip": true, "unequip": true,
			"use": true, "drop": true, "take": true,
		},
		chineseDirectionPattern: regexp.MustCompile(`^往?[北南东西西北东北西南东南出][走去]?$`),
		englishDirectionPattern: regexp.MustCompile(`^(go|walk)\s+(north|south|east|west|northwest|northeast|southwest|southeast|out)$`),
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

func (iv *InputValidator) ValidateMovementCommand(input string) ValidationResult {
	sanitized := strings.TrimSpace(input)
	sanitizedLower := strings.ToLower(sanitized)

	// Check Chinese format (supports: 北, 往北, 往北走, 往北去, 西北, 往西北, 往西北走, etc.)
	if iv.chineseDirectionPattern.MatchString(sanitized) {
		return ValidationResult{
			IsValid:   true,
			Sanitized: sanitized,
		}
	}

	// Check if it's a valid Chinese direction without the pattern (for raw directions like "北")
	// Extract direction by removing optional prefix/suffix
	cleaned := sanitized
	cleaned = strings.TrimPrefix(cleaned, "往")
	cleaned = strings.TrimSuffix(cleaned, "走")
	cleaned = strings.TrimSuffix(cleaned, "去")

	validDirections := []string{"北", "南", "东", "西", "西北", "东北", "西南", "东南", "出"}
	for _, dir := range validDirections {
		if cleaned == dir {
			return ValidationResult{
				IsValid:   true,
				Sanitized: "往" + dir + "走", // Normalize format
			}
		}
	}

	// Check English format
	if strings.HasPrefix(sanitizedLower, "go ") || strings.HasPrefix(sanitizedLower, "walk ") {
		if iv.englishDirectionPattern.MatchString(sanitizedLower) {
			return ValidationResult{
				IsValid:   true,
				Sanitized: sanitized,
			}
		}

		return ValidationResult{
			IsValid:         false,
			Sanitized:       sanitized,
			ErrorMsg:        "Invalid direction. Please use: go north/south/east/west/northwest/northeast/southwest/southeast/out",
			ErrorMsgPinyin:  "Wu xiao fang xiang. Qing shi yong: go north/south/east/west/northwest/northeast/southwest/southeast/out",
			ErrorMsgEnglish: "Invalid direction. Please use: go north/south/east/west/northwest/northeast/southwest/southeast/out",
		}
	}

	// Check if it looks like a direction but missing proper format
	cleanedLower := strings.ToLower(cleaned)
	englishDirections := []string{"north", "south", "east", "west", "northwest", "northeast", "southwest", "southeast", "out"}
	for _, dir := range englishDirections {
		if cleanedLower == dir {
			return ValidationResult{
				IsValid:         false,
				Sanitized:       sanitized,
				ErrorMsg:        "请输入 'go " + dir + "' 或 '往" + getChineseDirection(dir) + "走'",
				ErrorMsgPinyin:  "Qing shu ru 'go " + dir + "' huo 'wang " + getChineseDirection(dir) + " zou'",
				ErrorMsgEnglish: "Please use 'go " + dir + "' or '往" + getChineseDirection(dir) + "走'",
			}
		}
	}

	return ValidationResult{
		IsValid:         false,
		Sanitized:       sanitized,
		ErrorMsg:        "未知命令",
		ErrorMsgPinyin:  "Wei zhi ming ling",
		ErrorMsgEnglish: "Unknown command",
	}
}

// Helper function to get Chinese direction from English
func getChineseDirection(english string) string {
	switch english {
	case "north":
		return "北"
	case "south":
		return "南"
	case "east":
		return "东"
	case "west":
		return "西"
	case "northwest":
		return "西北"
	case "northeast":
		return "东北"
	case "southwest":
		return "西南"
	case "southeast":
		return "东南"
	case "out":
		return "出"
	default:
		return english
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

	// Check quit commands first
	if sanitized == "quit" || sanitized == "exit" || sanitized == "退出" || sanitized == "退" {
		return "quit"
	}

	switch sanitized {
	case "help", "帮助":
		return "help"
	case "inventory", "i", "背包":
		return "inventory"
	case "status", "状态":
		return "status"
	case "look", "看", "查看":
		return "look"
	case "saves", "存档列表":
		return "list_saves"
	default:
		// Check for movement
		cleaned := strings.TrimPrefix(sanitized, "往")
		cleaned = strings.TrimSuffix(cleaned, "走")
		cleaned = strings.TrimSuffix(cleaned, "去")

		validDirections := []string{"北", "南", "东", "西", "西北", "东北", "西南", "东南", "出"}
		for _, dir := range validDirections {
			if strings.Contains(sanitized, dir) {
				return "movement"
			}
		}

		// Check for movement (English)
		if strings.HasPrefix(sanitized, "go ") || strings.HasPrefix(sanitized, "walk ") {
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
