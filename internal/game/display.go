package game

import (
	"fmt"
	"strings"

	"golden_chalice/internal/models"
)

// DisplayFormatter handles all text formatting for the game
type DisplayFormatter struct {
	config *models.Config
}

// NewDisplayFormatter creates a new display formatter
func NewDisplayFormatter(config *models.Config) *DisplayFormatter {
	return &DisplayFormatter{
		config: config,
	}
}

// FormatText formats a Text struct based on display preferences
func (df *DisplayFormatter) FormatText(text models.Text) string {
	var parts []string

	if df.config.ShouldShowChinese() && text.Chinese != "" {
		parts = append(parts, text.Chinese)
	}

	if df.config.ShouldShowPinyin() && text.Pinyin != "" {
		pinyinText := text.Pinyin
		if df.config.ShouldShowChinese() && len(parts) > 0 {
			pinyinText = "(" + pinyinText + ")"
		}
		parts = append(parts, pinyinText)
	}

	if df.config.ShouldShowEnglish() && text.English != "" {
		englishText := text.English
		if len(parts) > 0 {
			englishText = "/ " + englishText
		}
		parts = append(parts, englishText)
	}

	if len(parts) == 0 {
		return text.Chinese
	}

	return strings.Join(parts, " ")
}

// ShowChinese returns whether Chinese text should be displayed
func (df *DisplayFormatter) ShowChinese() bool {
	return df.config.ShouldShowChinese()
}

// ShowPinyin returns whether Pinyin text should be displayed
func (df *DisplayFormatter) ShowPinyin() bool {
	return df.config.ShouldShowPinyin()
}

// ShowEnglish returns whether English text should be displayed
func (df *DisplayFormatter) ShowEnglish() bool {
	return df.config.ShouldShowEnglish()
}

// FormatName is a convenience method for formatting names
func (df *DisplayFormatter) FormatName(name models.Text) string {
	return df.FormatText(name)
}

// FormatDescription is a convenience method for formatting descriptions
func (df *DisplayFormatter) FormatDescription(desc models.Text) string {
	return df.FormatText(desc)
}

// FormatHealthBar creates a visual health bar
func (df *DisplayFormatter) FormatHealthBar(current, max int, width int) string {
	if max <= 0 {
		return ""
	}

	percentage := float64(current) / float64(max)
	filled := int(percentage * float64(width))

	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	empty := width - filled

	bar := "["
	bar += strings.Repeat("█", filled)
	bar += strings.Repeat("░", empty)
	bar += "]"

	percent := int(percentage * 100)
	bar += fmt.Sprintf(" %d%%", percent)

	return bar
}

// FormatHealthStatus returns a colored health status indicator
func (df *DisplayFormatter) FormatHealthStatus(current, max int) string {
	percentage := float64(current) / float64(max) * 100

	var chinese, pinyin, english string

	switch {
	case percentage >= 75:
		chinese = "健康"
		pinyin = "Jian Kang"
		english = "Healthy"
	case percentage >= 50:
		chinese = "轻伤"
		pinyin = "Qing Shang"
		english = "Injured"
	case percentage >= 25:
		chinese = "重伤"
		pinyin = "Zhong Shang"
		english = "Severely Injured"
	case percentage > 0:
		chinese = "濒死"
		pinyin = "Bin Si"
		english = "Dying"
	default:
		chinese = "死亡"
		pinyin = "Si Wang"
		english = "Dead"
	}

	return df.formatInline(chinese, pinyin, english)
}

// FormatDamage formats damage numbers
func (df *DisplayFormatter) FormatDamage(amount int) string {
	chinese := fmt.Sprintf("-%d", amount)
	pinyin := fmt.Sprintf("-%d", amount)
	english := fmt.Sprintf("-%d", amount)
	return df.formatInline(chinese, pinyin, english)
}

// FormatHeal formats heal numbers
func (df *DisplayFormatter) FormatHeal(amount int) string {
	chinese := fmt.Sprintf("+%d", amount)
	pinyin := fmt.Sprintf("+%d", amount)
	english := fmt.Sprintf("+%d", amount)
	return df.formatInline(chinese, pinyin, english)
}

// FormatExperience formats experience gain
func (df *DisplayFormatter) FormatExperience(amount int) string {
	chinese := fmt.Sprintf("+%d 经验值", amount)
	pinyin := fmt.Sprintf("+%d jing yan zhi", amount)
	english := fmt.Sprintf("+%d XP", amount)
	return df.formatInline(chinese, pinyin, english)
}

// FormatLevelUp formats level up message
func (df *DisplayFormatter) FormatLevelUp(level int) string {
	chinese := fmt.Sprintf("🎉 升级！当前等级: %d 🎉", level)
	pinyin := fmt.Sprintf("🎉 Sheng Ji! Dang Qian Deng Ji: %d 🎉", level)
	english := fmt.Sprintf("🎉 Level Up! Current level: %d 🎉", level)
	return df.formatInline(chinese, pinyin, english)
}

// FormatCombatMessage formats combat messages with consistent styling
func (df *DisplayFormatter) FormatCombatMessage(attacker, target, action string, damage int) string {
	var chinese, pinyin, english string

	if damage > 0 {
		chinese = fmt.Sprintf("%s %s %s，造成 %d 点伤害！", attacker, action, target, damage)
		pinyin = fmt.Sprintf("%s %s %s，zao cheng %d dian shang hai!", attacker, action, target, damage)
		english = fmt.Sprintf("%s %s %s for %d damage!", attacker, action, target, damage)
	} else {
		chinese = fmt.Sprintf("%s %s %s，但没有造成伤害！", attacker, action, target)
		pinyin = fmt.Sprintf("%s %s %s，dan mei you zao cheng shang hai!", attacker, action, target)
		english = fmt.Sprintf("%s %s %s but deals no damage!", attacker, action, target)
	}

	return df.formatInline(chinese, pinyin, english)
}

// FormatVictory formats victory message
func (df *DisplayFormatter) FormatVictory() string {
	chinese := "✨✨✨ 胜利！ ✨✨✨"
	pinyin := "✨✨✨ Sheng Li! ✨✨✨"
	english := "✨✨✨ Victory! ✨✨✨"
	return df.formatInline(chinese, pinyin, english)
}

// FormatGameOver formats game over message
func (df *DisplayFormatter) FormatGameOver() string {
	chinese := "☠️ 游戏结束 ☠️"
	pinyin := "☠️ You Xi Jie Shu ☠️"
	english := "☠️ Game Over ☠️"
	return df.formatInline(chinese, pinyin, english)
}

// FormatItemPickup formats item pickup message
func (df *DisplayFormatter) FormatItemPickup(itemName models.Text) string {
	chinese := fmt.Sprintf("你拿起了 %s", itemName.Chinese)
	pinyin := fmt.Sprintf("Ni na qi le %s", itemName.Pinyin)
	english := fmt.Sprintf("You picked up: %s", itemName.English)
	return df.formatInline(chinese, pinyin, english)
}

// FormatItemDrop formats item drop message
func (df *DisplayFormatter) FormatItemDrop(itemName models.Text) string {
	chinese := fmt.Sprintf("你丢弃了 %s", itemName.Chinese)
	pinyin := fmt.Sprintf("Niu diu qi le %s", itemName.Pinyin)
	english := fmt.Sprintf("You dropped: %s", itemName.English)
	return df.formatInline(chinese, pinyin, english)
}

// FormatItemUse formats item use message
func (df *DisplayFormatter) FormatItemUse(itemName models.Text, effect string) string {
	chinese := fmt.Sprintf("你使用了 %s，%s", itemName.Chinese, effect)
	pinyin := fmt.Sprintf("Ni shi yong le %s，%s", itemName.Pinyin, effect)
	english := fmt.Sprintf("You used %s, %s", itemName.English, effect)
	return df.formatInline(chinese, pinyin, english)
}

// FormatEquipment formats equipment message
func (df *DisplayFormatter) FormatEquipment(itemName models.Text, action string) string {
	var chinese, pinyin, english string

	if action == "equip" {
		chinese = fmt.Sprintf("装备了 %s", itemName.Chinese)
		pinyin = fmt.Sprintf("Zhuang bei le %s", itemName.Pinyin)
		english = fmt.Sprintf("Equipped: %s", itemName.English)
	} else {
		chinese = fmt.Sprintf("卸下了 %s", itemName.Chinese)
		pinyin = fmt.Sprintf("Xie xia le %s", itemName.Pinyin)
		english = fmt.Sprintf("Unequipped: %s", itemName.English)
	}

	return df.formatInline(chinese, pinyin, english)
}

// FormatRoomTitle formats a room title with decorations
func (df *DisplayFormatter) FormatRoomTitle(name models.Text) string {
	formattedName := df.FormatName(name)
	return fmt.Sprintf("\n=== %s ===\n", formattedName)
}

// FormatSeparator returns a decorative separator line
func (df *DisplayFormatter) FormatSeparator() string {
	return "----------------------------------------"
}

// FormatHeader formats a section header
func (df *DisplayFormatter) FormatHeader(title models.Text) string {
	formattedTitle := df.FormatName(title)
	return fmt.Sprintf("\n=== %s ===\n", formattedTitle)
}

// FormatSubHeader formats a subsection header
func (df *DisplayFormatter) FormatSubHeader(title models.Text) string {
	formattedTitle := df.FormatName(title)
	return fmt.Sprintf("--- %s ---\n", formattedTitle)
}

// FormatList formats a list of items with bullet points
func (df *DisplayFormatter) FormatList(items []string) string {
	if len(items) == 0 {
		return ""
	}

	var result strings.Builder
	for _, item := range items {
		result.WriteString(fmt.Sprintf("  • %s\n", item))
	}
	return result.String()
}

// FormatExitList formats a list of available exits
func (df *DisplayFormatter) FormatExitList(exits []models.Direction) string {
	if len(exits) == 0 {
		return ""
	}

	exitStrings := make([]string, len(exits))
	for i, exit := range exits {
		exitStrings[i] = exit.String()
	}

	chinese := "出口: " + strings.Join(exitStrings, ", ")
	pinyin := "Chu Kou: " + strings.Join(exitStrings, ", ")
	english := "Exits: " + strings.Join(exitStrings, ", ")

	return df.formatInline(chinese, pinyin, english)
}

// FormatItemList formats a list of already-formatted item names
// Deprecated: Use FormatItemListFromTexts instead
func (df *DisplayFormatter) FormatItemList(itemNames []string) string {
	if len(itemNames) == 0 {
		return ""
	}

	chinese := "你可以看到: " + strings.Join(itemNames, ", ")
	pinyin := "Ni ke yi kan dao: " + strings.Join(itemNames, ", ")
	english := "You see: " + strings.Join(itemNames, ", ")

	return df.formatInline(chinese, pinyin, english)
}

// FormatItemListFromTexts formats a list of items from Text structs
// This is the preferred method for displaying item lists
func (df *DisplayFormatter) FormatItemListFromTexts(items []models.Text) string {
	if len(items) == 0 {
		return ""
	}

	// Format each item individually using the display formatter
	formattedItems := make([]string, len(items))
	for i, item := range items {
		formattedItems[i] = df.FormatText(item)
	}

	// Join the formatted items with commas
	joinedItems := strings.Join(formattedItems, ", ")

	// Create the full trilingual line
	chinese := "你可以看到: " + joinedItems
	pinyin := "Ni ke yi kan dao: " + joinedItems
	english := "You see: " + joinedItems

	return df.formatInline(chinese, pinyin, english)
}

// FormatSimpleItemList formats a list of already-formatted item names
// This method does NOT re-format the item names - they should already be formatted
func (df *DisplayFormatter) FormatSimpleItemList(formattedItemNames []string) string {
	if len(formattedItemNames) == 0 {
		return ""
	}

	// Join the already-formatted item names with commas
	joinedItems := strings.Join(formattedItemNames, ", ")

	// Create the trilingual prefix only (the items are already trilingual)
	chinese := "你可以看到: " + joinedItems
	pinyin := "Ni ke yi kan dao: " + joinedItems
	english := "You see: " + joinedItems

	return df.formatInline(chinese, pinyin, english)
}

// formatInline is an internal helper for simple trilingual strings
func (df *DisplayFormatter) formatInline(chinese, pinyin, english string) string {
	var parts []string

	if df.config.ShouldShowChinese() && chinese != "" {
		parts = append(parts, chinese)
	}

	if df.config.ShouldShowPinyin() && pinyin != "" {
		pinyinText := pinyin
		if df.config.ShouldShowChinese() && len(parts) > 0 {
			pinyinText = "(" + pinyinText + ")"
		}
		parts = append(parts, pinyinText)
	}

	if df.config.ShouldShowEnglish() && english != "" {
		englishText := english
		if len(parts) > 0 {
			englishText = "/ " + englishText
		}
		parts = append(parts, englishText)
	}

	if len(parts) == 0 {
		return chinese
	}

	return strings.Join(parts, " ")
}

// FormatTwoColumn formats two columns of text (e.g., for status display)
func (df *DisplayFormatter) FormatTwoColumn(label, value string) string {
	return fmt.Sprintf("%-15s: %s", label, value)
}

// FormatProgress formats a progress indicator for experience or health
func (df *DisplayFormatter) FormatProgress(current, max int, width int) string {
	return df.FormatHealthBar(current, max, width)
}

// FormatWelcome formats the welcome message
func (df *DisplayFormatter) FormatWelcome(version string) string {
	var parts []string

	// Chinese version
	if df.config.ShouldShowChinese() {
		parts = append(parts, fmt.Sprintf(`╔═══════════════════════════════════════════════════════════╗
║              黑暗洞穴 - 文字冒险游戏                      ║
║                    版本: %-33s║
╚═══════════════════════════════════════════════════════════╝`, version))
	}

	// Pinyin version
	if df.config.ShouldShowPinyin() {
		parts = append(parts, fmt.Sprintf(`╔═══════════════════════════════════════════════════════════╗
║             Hei An Dong Xue - Wen Zi Mao Xian You Xi      ║
║                     Ban Ben: %-29s║
╚═══════════════════════════════════════════════════════════╝`, version))
	}

	// English version
	if df.config.ShouldShowEnglish() {
		parts = append(parts, fmt.Sprintf(`╔═══════════════════════════════════════════════════════════╗
║             Dark Caverns - Text Adventure RPG             ║
║                     Version: %-29s║
╚═══════════════════════════════════════════════════════════╝`, version))
	}

	return strings.Join(parts, "\n")
}
