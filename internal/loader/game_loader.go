package loader

import (
	"fmt"
	"path/filepath"

	"golden_chalice/internal/models"
)

// LoadedGameData contains all loaded game data from JSON files
type LoadedGameData struct {
	Config  *models.Config
	Biomes  map[string]models.Biome
	Items   map[string]models.Item
	Enemies map[string]models.Enemy
	World   *models.World
}

// LoadAllGameData loads and validates all game data from JSON files
// This is the main entry point for game initialization
func LoadAllGameData(dataDir string) (*LoadedGameData, error) {
	// 1. Load config first (no dependencies)
	config, err := LoadConfig(filepath.Join(dataDir, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	fmt.Println("✓ Config loaded successfully")

	// 2. Load biomes (no dependencies)
	biomes, err := LoadBiomes(filepath.Join(dataDir, "biomes.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to load biomes: %w", err)
	}
	fmt.Printf("✓ Biomes loaded: %d biomes\n", len(biomes))

	// 3. Load items (no dependencies)
	items, err := LoadItems(filepath.Join(dataDir, "items.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to load items: %w", err)
	}
	fmt.Printf("✓ Items loaded: %d items\n", len(items))

	// 4. Load enemies with item validation (depends on items)
	enemies, err := LoadEnemiesWithItemValidation(filepath.Join(dataDir, "enemies.json"), items)
	if err != nil {
		return nil, fmt.Errorf("failed to load enemies: %w", err)
	}
	fmt.Printf("✓ Enemies loaded: %d enemies\n", len(enemies))

	// 5. Load world with full validation (depends on biomes, enemies, items)
	world, err := LoadWorldWithValidation(filepath.Join(dataDir, "world.json"), biomes, enemies, items)
	if err != nil {
		return nil, fmt.Errorf("failed to load world: %w", err)
	}
	fmt.Printf("✓ World loaded: %d locations\n", len(world.Locations))

	// 6. Additional cross-validation: Verify win condition item exists
	if err := validateWinCondition(config, items); err != nil {
		return nil, err
	}

	// 7. Additional cross-validation: Verify starting location exists
	if err := validateStartingLocation(config, world); err != nil {
		return nil, err
	}

	fmt.Println("\n✓ All game data loaded and validated successfully!")
	return &LoadedGameData{
		Config:  config,
		Biomes:  biomes,
		Items:   items,
		Enemies: enemies,
		World:   world,
	}, nil
}

// validateWinCondition ensures the win condition item exists in items.json
func validateWinCondition(config *models.Config, items map[string]models.Item) error {
	winItemID := config.GetWinConditionItemID()
	if _, exists := items[winItemID]; !exists {
		return fmt.Errorf("win condition item '%s' not found in items.json", winItemID)
	}

	// Verify it's actually a quest item (optional but helpful)
	item, _ := GetItem(items, winItemID)
	if !item.IsQuestItem() && !item.IsWinCondition() {
		fmt.Printf("WARNING: Win condition item '%s' is not marked as a quest item or win condition\n", winItemID)
	}

	fmt.Printf("✓ Win condition item validated: %s\n", winItemID)
	return nil
}

// validateStartingLocation ensures the starting location exists in world.json
func validateStartingLocation(config *models.Config, world *models.World) error {
	startID := config.GetStartingLocationID()
	if _, err := world.GetLocation(startID); err != nil {
		return fmt.Errorf("starting location '%s' not found in world.json: %w", startID, err)
	}

	fmt.Printf("✓ Starting location validated: %s\n", startID)
	return nil
}

// QuickLoad is a convenience function for development that loads with default paths
// Assumes data files are in "./data" directory
func QuickLoad() (*LoadedGameData, error) {
	return LoadAllGameData("./data")
}

// MustLoad loads game data and panics on error (useful for development)
func MustLoad(dataDir string) *LoadedGameData {
	data, err := LoadAllGameData(dataDir)
	if err != nil {
		panic(fmt.Sprintf("Failed to load game data: %v", err))
	}
	return data
}

// GetLocationName returns the formatted name of a location using config preferences
func (d *LoadedGameData) GetLocationName(locationID string) string {
	location, err := d.World.GetLocation(locationID)
	if err != nil {
		return locationID
	}
	return d.Config.GetDisplayText(location.Name)
}

// GetItemName returns the formatted name of an item using config preferences
func (d *LoadedGameData) GetItemName(itemID string) string {
	item, err := GetItem(d.Items, itemID)
	if err != nil {
		return itemID
	}
	return d.Config.GetDisplayText(item.Name)
}

// GetEnemyName returns the formatted name of an enemy using config preferences
func (d *LoadedGameData) GetEnemyName(enemyID string) string {
	enemy, err := GetEnemy(d.Enemies, enemyID)
	if err != nil {
		return enemyID
	}
	return d.Config.GetDisplayText(enemy.Name)
}

// GetBiomeName returns the formatted name of a biome using config preferences
func (d *LoadedGameData) GetBiomeName(biomeID string) string {
	biome, err := GetBiome(d.Biomes, biomeID)
	if err != nil {
		return biomeID
	}
	return d.Config.GetDisplayText(biome.Name)
}
