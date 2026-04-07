package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golden_chalice/internal/models"
)

// LoadConfig reads and parses the config.json file
func LoadConfig(filePath string) (*models.Config, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Unmarshal JSON
	var config models.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	// Validate the configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validateConfig checks that all configuration values are valid
func validateConfig(config *models.Config) error {
	var errors []string

	// Check game version
	if config.GameVersion == "" {
		errors = append(errors, "game_version is required")
	}

	// Check starting location
	if config.StartingLocationID == "" {
		errors = append(errors, "starting_location_id is required")
	}

	// Check win condition item
	if config.WinConditionItemID == "" {
		errors = append(errors, "win_condition_item_id is required")
	}

	// Check starting health
	if config.StartingHealth <= 0 {
		errors = append(errors, fmt.Sprintf("starting_health must be positive, got %d", config.StartingHealth))
	}

	// Check display preferences (at least one should be true)
	if !config.DisplayPreferences.ShowChinese &&
		!config.DisplayPreferences.ShowPinyin &&
		!config.DisplayPreferences.ShowEnglish {
		errors = append(errors, "at least one display preference (chinese, pinyin, english) must be enabled")
	}

	// Check combat settings
	if config.CombatSettings.BaseDamageMin < 0 {
		errors = append(errors, fmt.Sprintf("base_damage_min cannot be negative, got %d", config.CombatSettings.BaseDamageMin))
	}

	if config.CombatSettings.BaseDamageMax < 0 {
		errors = append(errors, fmt.Sprintf("base_damage_max cannot be negative, got %d", config.CombatSettings.BaseDamageMax))
	}

	if config.CombatSettings.BaseDamageMin > config.CombatSettings.BaseDamageMax {
		errors = append(errors, fmt.Sprintf("base_damage_min (%d) > base_damage_max (%d)",
			config.CombatSettings.BaseDamageMin, config.CombatSettings.BaseDamageMax))
	}

	if config.CombatSettings.FleeSuccessRate < 0 || config.CombatSettings.FleeSuccessRate > 100 {
		errors = append(errors, fmt.Sprintf("flee_success_rate must be between 0 and 100, got %d",
			config.CombatSettings.FleeSuccessRate))
	}

	if config.CombatSettings.PlayerBaseDefense < 0 {
		errors = append(errors, fmt.Sprintf("player_base_defense cannot be negative, got %d",
			config.CombatSettings.PlayerBaseDefense))
	}

	if len(errors) > 0 {
		return fmt.Errorf("found %d config validation errors:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}
