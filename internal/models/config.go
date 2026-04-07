package models

// DisplayPreferences controls what information is shown to the player
type DisplayPreferences struct {
	ShowChinese bool `json:"show_chinese"`
	ShowPinyin  bool `json:"show_pinyin"`
	ShowEnglish bool `json:"show_english"`
}

// CombatSettings contains game balance configuration
type CombatSettings struct {
	BaseDamageMin     int `json:"base_damage_min"`     // Minimum player base damage
	BaseDamageMax     int `json:"base_damage_max"`     // Maximum player base damage
	FleeSuccessRate   int `json:"flee_success_rate"`   // Percentage chance to flee (0-100)
	PlayerBaseDefense int `json:"player_base_defense"` // Base defense before equipment
}

// Config represents the entire game configuration
type Config struct {
	GameVersion        string             `json:"game_version"`
	StartingLocationID string             `json:"starting_location_id"`
	WinConditionItemID string             `json:"win_condition_item_id"`
	StartingHealth     int                `json:"starting_health"`
	DisplayPreferences DisplayPreferences `json:"display_preferences"`
	CombatSettings     CombatSettings     `json:"combat_settings"`
}

// GetDisplayText formats a Text struct based on display preferences
// Returns a formatted string combining enabled languages
func (c *Config) GetDisplayText(text Text) string {
	var parts []string

	if c.DisplayPreferences.ShowChinese && text.Chinese != "" {
		parts = append(parts, text.Chinese)
	}

	if c.DisplayPreferences.ShowPinyin && text.Pinyin != "" {
		pinyinText := text.Pinyin
		// Only add parentheses if showing Chinese as well
		if c.DisplayPreferences.ShowChinese && len(parts) > 0 {
			pinyinText = "(" + pinyinText + ")"
		}
		parts = append(parts, pinyinText)
	}

	if c.DisplayPreferences.ShowEnglish && text.English != "" {
		englishText := text.English
		// Add separator if showing other languages
		if len(parts) > 0 {
			englishText = "/ " + englishText
		}
		parts = append(parts, englishText)
	}

	if len(parts) == 0 {
		// Fallback to Chinese if available, otherwise empty string
		return text.Chinese
	}

	// Join with spaces
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " "
		}
		result += part
	}
	return result
}

// GetFormattedName returns a formatted location or item name
func (c *Config) GetFormattedName(name Text) string {
	return c.GetDisplayText(name)
}

// GetFormattedDescription returns a formatted description
func (c *Config) GetFormattedDescription(description Text) string {
	return c.GetDisplayText(description)
}

// GetCombatDamageRange returns the min and max damage for player attacks
func (c *Config) GetCombatDamageRange() (min, max int) {
	return c.CombatSettings.BaseDamageMin, c.CombatSettings.BaseDamageMax
}

// GetFleeSuccessRate returns the flee success rate as a percentage
func (c *Config) GetFleeSuccessRate() int {
	return c.CombatSettings.FleeSuccessRate
}

// GetPlayerBaseDefense returns the player's base defense value
func (c *Config) GetPlayerBaseDefense() int {
	return c.CombatSettings.PlayerBaseDefense
}

// ShouldShowChinese returns whether Chinese text should be displayed
func (c *Config) ShouldShowChinese() bool {
	return c.DisplayPreferences.ShowChinese
}

// ShouldShowPinyin returns whether Pinyin text should be displayed
func (c *Config) ShouldShowPinyin() bool {
	return c.DisplayPreferences.ShowPinyin
}

// ShouldShowEnglish returns whether English text should be displayed
func (c *Config) ShouldShowEnglish() bool {
	return c.DisplayPreferences.ShowEnglish
}

// SetDisplayPreferences updates display preferences (useful for in-game options)
func (c *Config) SetDisplayPreferences(showChinese, showPinyin, showEnglish bool) {
	c.DisplayPreferences.ShowChinese = showChinese
	c.DisplayPreferences.ShowPinyin = showPinyin
	c.DisplayPreferences.ShowEnglish = showEnglish

	// Ensure at least one is true
	if !showChinese && !showPinyin && !showEnglish {
		c.DisplayPreferences.ShowChinese = true // Fallback to Chinese
	}
}

// GetStartingHealth returns the player's starting health
func (c *Config) GetStartingHealth() int {
	return c.StartingHealth
}

// GetStartingLocationID returns the starting location ID
func (c *Config) GetStartingLocationID() string {
	return c.StartingLocationID
}

// GetWinConditionItemID returns the win condition item ID
func (c *Config) GetWinConditionItemID() string {
	return c.WinConditionItemID
}

// GetGameVersion returns the game version string
func (c *Config) GetGameVersion() string {
	return c.GameVersion
}
