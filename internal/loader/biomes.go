package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golden_chalice/internal/models"
)

// LoadBiomes reads and parses the biomes.json file
func LoadBiomes(filePath string) (map[string]models.Biome, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read biomes file %s: %w", filePath, err)
	}

	// Unmarshal JSON into a struct with a biomes array
	var biomesData struct {
		Biomes []models.Biome `json:"biomes"`
	}

	if err := json.Unmarshal(data, &biomesData); err != nil {
		return nil, fmt.Errorf("failed to parse biomes.json: %w", err)
	}

	// Convert to map for O(1) lookup
	biomesMap := make(map[string]models.Biome)
	for _, biome := range biomesData.Biomes {
		// Check for duplicate IDs
		if _, exists := biomesMap[biome.ID]; exists {
			return nil, fmt.Errorf("duplicate biome ID found: %s", biome.ID)
		}
		biomesMap[biome.ID] = biome
	}

	// Validate the biomes
	if err := validateBiomes(biomesMap); err != nil {
		return nil, err
	}

	return biomesMap, nil
}

// validateBiomes performs validation on loaded biome data
func validateBiomes(biomes map[string]models.Biome) error {
	var errors []string

	for id, biome := range biomes {
		// Check ID matches map key
		if biome.ID != id {
			errors = append(errors, fmt.Sprintf("biome map key '%s' does not match biome.ID '%s'", id, biome.ID))
		}

		// Validate Text fields have content
		if biome.Name.Chinese == "" {
			errors = append(errors, fmt.Sprintf("biome '%s' has empty Chinese name", id))
		}
		if biome.Name.Pinyin == "" {
			errors = append(errors, fmt.Sprintf("biome '%s' has empty Pinyin name", id))
		}
		if biome.Name.English == "" {
			errors = append(errors, fmt.Sprintf("biome '%s' has empty English name", id))
		}

		// Ambient description validation (optional, but warn if completely missing)
		if biome.AmbientDescription.Chinese == "" && biome.AmbientDescription.Pinyin == "" && biome.AmbientDescription.English == "" {
			// Not a fatal error, just a warning
			fmt.Printf("WARNING: biome '%s' has no ambient description\n", id)
		}

		// Validate environmental effects if present
		if biome.EnvironmentalEffects != nil {
			effects := biome.EnvironmentalEffects

			// Validate modifier ranges (if not nil)
			if effects.AccuracyModifier != nil {
				if *effects.AccuracyModifier < 0 || *effects.AccuracyModifier > 2 {
					errors = append(errors, fmt.Sprintf("biome '%s' has invalid accuracy_modifier: %f (must be between 0 and 2)", id, *effects.AccuracyModifier))
				}
			}

			if effects.DodgeModifier != nil {
				if *effects.DodgeModifier < 0 || *effects.DodgeModifier > 2 {
					errors = append(errors, fmt.Sprintf("biome '%s' has invalid dodge_modifier: %f (must be between 0 and 2)", id, *effects.DodgeModifier))
				}
			}

			if effects.DamageModifier != nil {
				if *effects.DamageModifier < 0 || *effects.DamageModifier > 2 {
					errors = append(errors, fmt.Sprintf("biome '%s' has invalid damage_modifier: %f (must be between 0 and 2)", id, *effects.DamageModifier))
				}
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("found %d biome validation errors:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}

// GetBiome retrieves a biome by ID from the map
func GetBiome(biomes map[string]models.Biome, id string) (models.Biome, error) {
	biome, exists := biomes[id]
	if !exists {
		return models.Biome{}, fmt.Errorf("biome '%s' not found", id)
	}
	return biome, nil
}
