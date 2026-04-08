package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golden_chalice/internal/models"
)

// LoadItems reads and parses the items.json file
func LoadItems(filePath string) (map[string]models.Item, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read items file %s: %w", filePath, err)
	}

	// Unmarshal JSON into a struct with an items array
	var itemsData struct {
		Items []models.Item `json:"items"`
	}

	if err := json.Unmarshal(data, &itemsData); err != nil {
		return nil, fmt.Errorf("failed to parse items.json: %w", err)
	}

	// Convert to map for O(1) lookup
	itemsMap := make(map[string]models.Item)
	for _, item := range itemsData.Items {
		// Check for duplicate IDs
		if _, exists := itemsMap[item.ID]; exists {
			return nil, fmt.Errorf("duplicate item ID found: %s", item.ID)
		}
		itemsMap[item.ID] = item
	}

	// Validate items
	if err := validateItems(itemsMap); err != nil {
		return nil, err
	}

	return itemsMap, nil
}

// validateItems performs validation on loaded item data
func validateItems(items map[string]models.Item) error {
	var errors []string

	for id, item := range items {
		// Check ID matches map key
		if item.ID != id {
			errors = append(errors, fmt.Sprintf("item map key '%s' does not match item.ID '%s'", id, item.ID))
		}

		// Validate Text fields have content
		if item.Name.Chinese == "" {
			errors = append(errors, fmt.Sprintf("item '%s' has empty Chinese name", id))
		}
		if item.Name.Pinyin == "" {
			errors = append(errors, fmt.Sprintf("item '%s' has empty Pinyin name", id))
		}
		if item.Name.English == "" {
			errors = append(errors, fmt.Sprintf("item '%s' has empty English name", id))
		}

		// Validate description (recommended but not required)
		if item.Description.Chinese == "" && item.Description.Pinyin == "" && item.Description.English == "" {
			fmt.Printf("WARNING: item '%s' has no description\n", id)
		}

		// Validate item type
		switch item.Type {
		case models.ItemTypeWeapon, models.ItemTypeQuest, models.ItemTypeConsumable, models.ItemTypeKey, models.ItemTypeArmor, models.ItemTypeBackpack:
			// Valid type
		default:
			errors = append(errors, fmt.Sprintf("item '%s' has invalid type: %s", id, item.Type))
		}

		// Validate properties based on type
		if err := validateItemProperties(item); err != nil {
			errors = append(errors, fmt.Sprintf("item '%s' property error: %v", id, err))
		}

		// Validate consumable flag consistency
		if item.Consumable && !item.Usable {
			errors = append(errors, fmt.Sprintf("item '%s' is consumable but not usable", id))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("found %d item validation errors:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}

// validateItemProperties checks that properties match item type
func validateItemProperties(item models.Item) error {
	props := item.Properties

	switch item.Type {
	case models.ItemTypeWeapon:
		if props.DamageBonus == nil {
			return fmt.Errorf("weapon missing damage_bonus")
		}
		if props.DamageBonus != nil && *props.DamageBonus < 0 {
			return fmt.Errorf("weapon has negative damage_bonus: %d", *props.DamageBonus)
		}

	case models.ItemTypeConsumable:
		if props.HealthRestore == nil && props.ManaRestore == nil {
			return fmt.Errorf("consumable missing health_restore or mana_restore")
		}
		if props.HealthRestore != nil && *props.HealthRestore < 0 {
			return fmt.Errorf("consumable has negative health_restore: %d", *props.HealthRestore)
		}
		if props.ManaRestore != nil && *props.ManaRestore < 0 {
			return fmt.Errorf("consumable has negative mana_restore: %d", *props.ManaRestore)
		}

	case models.ItemTypeQuest:
		// Quest items can have win_condition or just story flags
		// No strict validation needed

	case models.ItemTypeKey:
		// Keys may or may not specify which door they open
		// No strict validation needed

	case models.ItemTypeArmor:
		if props.DefenseBonus == nil {
			return fmt.Errorf("armor missing defense_bonus")
		}
		if props.DefenseBonus != nil && *props.DefenseBonus < 0 {
			return fmt.Errorf("armor has negative defense_bonus: %d", *props.DefenseBonus)
		}

	case models.ItemTypeBackpack:
		// Backpack validation - size_bonus is required
		if props.SizeBonus == nil {
			return fmt.Errorf("backpack missing size_bonus")
		}
		if props.SizeBonus != nil && *props.SizeBonus < 0 {
			return fmt.Errorf("backpack has negative size_bonus: %d", *props.SizeBonus)
		}
		// Equippable should be true for backpacks
		if props.Equippable == nil || !*props.Equippable {
			return fmt.Errorf("backpack should be equippable")
		}
	}

	return nil
}

// GetItem retrieves an item by ID from the map
func GetItem(items map[string]models.Item, id string) (models.Item, error) {
	item, exists := items[id]
	if !exists {
		return models.Item{}, fmt.Errorf("item '%s' not found", id)
	}
	return item, nil
}

// GetItemsByType returns all items of a specific type
func GetItemsByType(items map[string]models.Item, itemType models.ItemType) []models.Item {
	var result []models.Item
	for _, item := range items {
		if item.Type == itemType {
			result = append(result, item)
		}
	}
	return result
}

// GetQuestItems returns all quest items
func GetQuestItems(items map[string]models.Item) []models.Item {
	return GetItemsByType(items, models.ItemTypeQuest)
}

// GetWeapons returns all weapons
func GetWeapons(items map[string]models.Item) []models.Item {
	return GetItemsByType(items, models.ItemTypeWeapon)
}

// GetConsumables returns all consumable items
func GetConsumables(items map[string]models.Item) []models.Item {
	return GetItemsByType(items, models.ItemTypeConsumable)
}

// GetKeys returns all keys
func GetKeys(items map[string]models.Item) []models.Item {
	return GetItemsByType(items, models.ItemTypeKey)
}

// GetArmor returns all armor items
func GetArmor(items map[string]models.Item) []models.Item {
	return GetItemsByType(items, models.ItemTypeArmor)
}

// ItemExists checks if an item ID exists in the items map
func ItemExists(items map[string]models.Item, id string) bool {
	_, exists := items[id]
	return exists
}
