package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golden_chalice/internal/models"
)

// LoadEnemies reads and parses the enemies.json file
func LoadEnemies(filePath string) (map[string]models.Enemy, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read enemies file %s: %w", filePath, err)
	}

	// Unmarshal JSON into a struct with an enemies array
	var enemiesData struct {
		Enemies []models.Enemy `json:"enemies"`
	}

	if err := json.Unmarshal(data, &enemiesData); err != nil {
		return nil, fmt.Errorf("failed to parse enemies.json: %w", err)
	}

	// Convert to map for O(1) lookup
	enemiesMap := make(map[string]models.Enemy)
	for _, enemy := range enemiesData.Enemies {
		// Check for duplicate IDs
		if _, exists := enemiesMap[enemy.ID]; exists {
			return nil, fmt.Errorf("duplicate enemy ID found: %s", enemy.ID)
		}
		enemiesMap[enemy.ID] = enemy
	}

	// Basic validation
	if err := validateEnemies(enemiesMap); err != nil {
		return nil, err
	}

	return enemiesMap, nil
}

// LoadEnemiesWithItemValidation loads enemies and validates drops against items map
func LoadEnemiesWithItemValidation(filePath string, items map[string]models.Item) (map[string]models.Enemy, error) {
	// Load enemies
	enemies, err := LoadEnemies(filePath)
	if err != nil {
		return nil, err
	}

	// Validate drops against items
	if err := validateEnemyDrops(enemies, items); err != nil {
		return nil, err
	}

	return enemies, nil
}

// validateEnemies performs basic validation on loaded enemy data
func validateEnemies(enemies map[string]models.Enemy) error {
	var errors []string

	for id, enemy := range enemies {
		// Check ID matches map key
		if enemy.ID != id {
			errors = append(errors, fmt.Sprintf("enemy map key '%s' does not match enemy.ID '%s'", id, enemy.ID))
		}

		// Validate Text fields have content
		if enemy.Name.Chinese == "" {
			errors = append(errors, fmt.Sprintf("enemy '%s' has empty Chinese name", id))
		}
		if enemy.Name.Pinyin == "" {
			errors = append(errors, fmt.Sprintf("enemy '%s' has empty Pinyin name", id))
		}
		if enemy.Name.English == "" {
			errors = append(errors, fmt.Sprintf("enemy '%s' has empty English name", id))
		}

		// Validate health
		if enemy.Health <= 0 {
			errors = append(errors, fmt.Sprintf("enemy '%s' has invalid health: %d", id, enemy.Health))
		}

		// Validate attack range
		if enemy.AttackPower.Min < 0 || enemy.AttackPower.Max < 0 {
			errors = append(errors, fmt.Sprintf("enemy '%s' has negative attack values", id))
		}
		if enemy.AttackPower.Min > enemy.AttackPower.Max {
			errors = append(errors, fmt.Sprintf("enemy '%s' has min attack (%d) > max attack (%d)", id, enemy.AttackPower.Min, enemy.AttackPower.Max))
		}

		// Validate defense (can be zero, but not negative)
		if enemy.Defense < 0 {
			errors = append(errors, fmt.Sprintf("enemy '%s' has negative defense: %d", id, enemy.Defense))
		}

		// Validate experience points
		if enemy.ExperiencePoints < 0 {
			errors = append(errors, fmt.Sprintf("enemy '%s' has negative experience points: %d", id, enemy.ExperiencePoints))
		}

		// Validate drop chances
		for i, drop := range enemy.Drops {
			if drop.Chance < 0 || drop.Chance > 100 {
				errors = append(errors, fmt.Sprintf("enemy '%s' drop %d has invalid chance: %d (must be 0-100)", id, i, drop.Chance))
			}
			if drop.ItemID == "" {
				errors = append(errors, fmt.Sprintf("enemy '%s' drop %d has empty item ID", id, i))
			}
		}

		// Validate special abilities (if present)
		if enemy.SpecialAbilities != nil {
			if enemy.SpecialAbilities.ChanceToActivate != nil {
				chance := *enemy.SpecialAbilities.ChanceToActivate
				if chance < 0 || chance > 100 {
					errors = append(errors, fmt.Sprintf("enemy '%s' has invalid special ability chance: %d", id, chance))
				}
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("found %d enemy validation errors:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}

// validateEnemyDrops checks that all drop item IDs exist in items map
func validateEnemyDrops(enemies map[string]models.Enemy, items map[string]models.Item) error {
	var errors []string

	for id, enemy := range enemies {
		for i, drop := range enemy.Drops {
			if _, exists := items[drop.ItemID]; !exists {
				errors = append(errors, fmt.Sprintf("enemy '%s' drop %d references unknown item '%s'", id, i, drop.ItemID))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("found %d drop validation errors:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}

// GetEnemy retrieves an enemy by ID from the map
func GetEnemy(enemies map[string]models.Enemy, id string) (models.Enemy, error) {
	enemy, exists := enemies[id]
	if !exists {
		return models.Enemy{}, fmt.Errorf("enemy '%s' not found", id)
	}
	return enemy, nil
}

// GetEnemiesByBiome returns all enemies that appear in a given biome
func GetEnemiesByBiome(enemies map[string]models.Enemy, biomeID string) []models.Enemy {
	var result []models.Enemy
	for _, enemy := range enemies {
		if enemy.AppearsInBiome(biomeID) {
			result = append(result, enemy)
		}
	}
	return result
}
