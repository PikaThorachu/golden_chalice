package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golden_chalice/internal/models"
)

// LoadWorld reads and parses the world.json file
func LoadWorld(filePath string) (*models.World, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read world file %s: %w", filePath, err)
	}

	// Unmarshal JSON
	var world models.World
	if err := json.Unmarshal(data, &world); err != nil {
		return nil, fmt.Errorf("failed to parse world.json: %w", err)
	}

	// Initialize empty rooms map if nil
	if world.Rooms == nil {
		world.Rooms = make(map[string]models.Room)
	}

	return &world, nil
}

// LoadWorldWithValidation loads the world and validates against provided data maps
func LoadWorldWithValidation(
	filePath string,
	biomes map[string]models.Biome,
	enemies map[string]models.Enemy,
	items map[string]models.Item,
) (*models.World, error) {
	// Load the world
	world, err := LoadWorld(filePath)
	if err != nil {
		return nil, err
	}

	// Validate the world
	if err := validateWorld(world, biomes, enemies, items); err != nil {
		return nil, fmt.Errorf("world validation failed: %w", err)
	}

	return world, nil
}

// validateWorld checks the integrity of the world data
func validateWorld(
	world *models.World,
	biomes map[string]models.Biome,
	enemies map[string]models.Enemy,
	items map[string]models.Item,
) error {
	var errors []string

	// Validate rooms first
	for id, room := range world.Rooms {
		if room.ID != id {
			errors = append(errors, fmt.Sprintf("room map key '%s' does not match room.ID '%s'", id, room.ID))
		}
		// Room names should have content
		if room.Name.Chinese == "" && room.Name.English == "" {
			errors = append(errors, fmt.Sprintf("room '%s' has empty name", id))
		}
	}

	// Check each location
	for id, location := range world.Locations {
		// Check location ID matches map key
		if location.ID != id {
			errors = append(errors, fmt.Sprintf("location map key '%s' does not match location.ID '%s'", id, location.ID))
		}

		// Check room exists if referenced
		if location.RoomID != nil {
			if _, exists := world.Rooms[*location.RoomID]; !exists {
				errors = append(errors, fmt.Sprintf("location '%s' references unknown room '%s'", id, *location.RoomID))
			}
		}

		// Check biome exists
		if _, exists := biomes[location.BiomeID]; !exists {
			errors = append(errors, fmt.Sprintf("location '%s' references unknown biome '%s'", id, location.BiomeID))
		}

		// Check enemies exist
		for _, enemyID := range location.EnemyIDs {
			if _, exists := enemies[enemyID]; !exists {
				errors = append(errors, fmt.Sprintf("location '%s' references unknown enemy '%s'", id, enemyID))
			}
		}

		// Check items exist
		for _, itemID := range location.ItemIDs {
			if _, exists := items[itemID]; !exists {
				errors = append(errors, fmt.Sprintf("location '%s' references unknown item '%s'", id, itemID))
			}
		}

		// Check exits
		for i, exit := range location.Exits {
			// Check destination exists
			if _, exists := world.Locations[exit.DestinationID]; !exists {
				errors = append(errors, fmt.Sprintf("location '%s' exit %d references unknown destination '%s'", id, i, exit.DestinationID))
			}

			// Check required item exists (if not nil)
			if exit.RequiredItem != nil {
				if _, exists := items[*exit.RequiredItem]; !exists {
					errors = append(errors, fmt.Sprintf("location '%s' exit %d references unknown required item '%s'", id, i, *exit.RequiredItem))
				}
			}
		}
	}

	// Check bidirectional exits (warnings only, not fatal)
	checkBidirectionalExits(world)

	if len(errors) > 0 {
		return fmt.Errorf("found %d validation errors:\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}

// checkBidirectionalExits logs warnings for exits that don't have a return path
func checkBidirectionalExits(world *models.World) {
	for fromID, location := range world.Locations {
		for _, exit := range location.Exits {
			toID := exit.DestinationID

			// Get the destination location
			destLocation, exists := world.Locations[toID]
			if !exists {
				// This error is already caught in validateWorld, skip here
				continue
			}

			// Check if destination has an exit back to source
			hasReturn := false
			for _, returnExit := range destLocation.Exits {
				if returnExit.DestinationID == fromID {
					hasReturn = true
					break
				}
			}

			if !hasReturn {
				// Log as warning
				fmt.Printf("WARNING: Exit from '%s' to '%s' has no return path\n", fromID, toID)
			}
		}
	}
}
