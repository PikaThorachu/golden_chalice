package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golden_chalice/internal/models"
)

// SaveData represents the complete save file structure
type SaveData struct {
	// Metadata
	Version     string    `json:"version"`      // Save file format version
	SaveTime    time.Time `json:"save_time"`    // When the save was created
	PlayerName  string    `json:"player_name"`  // Player name for quick identification
	LocationID  string    `json:"location_id"`  // Current location for save preview
	SaveVersion string    `json:"save_version"` // Version of the save format

	// Game state
	Player          *models.Player      `json:"player"`
	DefeatedEnemies map[string]bool     `json:"defeated_enemies"`
	TakenItems      map[string]bool     `json:"taken_items"`
	PendingDrops    map[string][]string `json:"pending_drops"`

	// Game metadata
	GameVersion string `json:"game_version"`
}

// SaveInfo contains metadata about a save file (for listing)
type SaveInfo struct {
	SlotName     string    `json:"slot_name"`
	DisplayName  string    `json:"display_name"`
	PlayerName   string    `json:"player_name"`
	LocationID   string    `json:"location_id"`
	LocationName string    `json:"location_name"` // Can be filled by caller
	SaveTime     time.Time `json:"save_time"`
	GameVersion  string    `json:"game_version"`
	IsAutoSlot   bool      `json:"is_auto_slot"`
	IsQuickSlot  bool      `json:"is_quick_slot"`
}

// SaveManager handles save file operations
type SaveManager struct {
	SaveDirectory string
	Config        *SaveConfig
	CurrentSave   *SaveData
	autoSaveTimer *time.Timer
}

// NewSaveManager creates a new save manager with configuration
func NewSaveManager(saveDir string, configPath string) (*SaveManager, error) {
	// Load configuration
	config, err := LoadSaveConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Override save directory if specified in config
	if config.SaveDirectory != "" && config.SaveDirectory != saveDir {
		saveDir = config.SaveDirectory
	}

	// Check if saves are enabled
	if !config.Enabled {
		fmt.Println("Save functionality is disabled in configuration")
	}

	// Create save directory if it doesn't exist
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create save directory: %w", err)
	}

	sm := &SaveManager{
		SaveDirectory: saveDir,
		Config:        config,
		CurrentSave:   nil,
	}

	// Start auto-save timer if enabled
	if config.AutoSaveEnabled {
		sm.startAutoSaveTimer()
	}

	return sm, nil
}

// startAutoSaveTimer starts the auto-save timer
func (sm *SaveManager) startAutoSaveTimer() {
	if sm.autoSaveTimer != nil {
		sm.autoSaveTimer.Stop()
	}

	duration := time.Duration(sm.Config.AutoSaveInterval) * time.Minute
	sm.autoSaveTimer = time.AfterFunc(duration, func() {
		if sm.CurrentSave != nil && sm.Config.AutoSaveEnabled {
			// Auto-save will be triggered by the game loop
			fmt.Println("\n[自动保存] 游戏正在自动保存...")
		}
	})
}

// IsSaveEnabled returns whether save functionality is enabled
func (sm *SaveManager) IsSaveEnabled() bool {
	return sm.Config.Enabled
}

// GetValidSlotNames returns all valid slot names
func (sm *SaveManager) GetValidSlotNames() []string {
	var slots []string
	for _, slot := range sm.Config.NamedSlots {
		slots = append(slots, slot.Name)
	}
	return slots
}

// CanSaveToSlot checks if a slot can be saved to
func (sm *SaveManager) CanSaveToSlot(slotName string) bool {
	if !sm.Config.Enabled {
		return false
	}

	slot := sm.Config.GetNamedSlot(slotName)
	if slot != nil && slot.IsReadOnly {
		return false
	}

	return sm.Config.IsValidSlotName(slotName)
}

// CreateSave creates a save from current game state
func (sm *SaveManager) CreateSave(
	player *models.Player,
	defeatedEnemies map[string]bool,
	takenItems map[string]bool,
	pendingDrops map[string][]string,
	gameVersion string,
	playerName string,
) (*SaveData, error) {

	saveData := &SaveData{
		Version:         "1.0",
		SaveTime:        time.Now(),
		SaveVersion:     "1.0",
		PlayerName:      playerName,
		LocationID:      player.CurrentLocationID,
		Player:          player,
		DefeatedEnemies: defeatedEnemies,
		TakenItems:      takenItems,
		PendingDrops:    pendingDrops,
		GameVersion:     gameVersion,
	}

	sm.CurrentSave = saveData
	return saveData, nil
}

// SaveToFile saves the current game state to a file with config validation
func (sm *SaveManager) SaveToFile(slotName string) error {
	if !sm.Config.Enabled {
		return fmt.Errorf("save functionality is disabled")
	}

	if !sm.CanSaveToSlot(slotName) {
		return fmt.Errorf("cannot save to slot: %s", slotName)
	}

	if sm.CurrentSave == nil {
		return fmt.Errorf("no save data to write")
	}

	// Update save time
	sm.CurrentSave.SaveTime = time.Now()

	// Marshal to JSON
	data, err := json.MarshalIndent(sm.CurrentSave, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	// Create filename
	filename := fmt.Sprintf("%s.json", slotName)
	filePath := filepath.Join(sm.SaveDirectory, filename)

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	fmt.Printf("游戏已保存到: %s\n", slotName)
	return nil
}

// LoadFromFile loads a save from a file
func (sm *SaveManager) LoadFromFile(slotName string) (*SaveData, error) {
	filename := fmt.Sprintf("%s.json", slotName)
	filePath := filepath.Join(sm.SaveDirectory, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("save file '%s' not found", slotName)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	// Unmarshal JSON
	var saveData SaveData
	if err := json.Unmarshal(data, &saveData); err != nil {
		return nil, fmt.Errorf("failed to parse save file: %w", err)
	}

	// Validate save data
	if err := sm.validateSaveData(&saveData); err != nil {
		return nil, fmt.Errorf("save file validation failed: %w", err)
	}

	sm.CurrentSave = &saveData
	fmt.Printf("游戏已加载: %s\n", slotName)
	return &saveData, nil
}

// validateSaveData performs basic validation on loaded save data
func (sm *SaveManager) validateSaveData(save *SaveData) error {
	if save.Player == nil {
		return fmt.Errorf("save file contains no player data")
	}

	if save.Player.CurrentLocationID == "" {
		return fmt.Errorf("save file has invalid location ID")
	}

	if save.Player.Health <= 0 {
		return fmt.Errorf("save file has invalid player health")
	}

	if save.DefeatedEnemies == nil {
		save.DefeatedEnemies = make(map[string]bool)
	}

	if save.TakenItems == nil {
		save.TakenItems = make(map[string]bool)
	}

	if save.PendingDrops == nil {
		save.PendingDrops = make(map[string][]string)
	}

	return nil
}

// ListSaves returns all available save files with display information
func (sm *SaveManager) ListSaves() ([]SaveInfo, error) {
	files, err := os.ReadDir(sm.SaveDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read save directory: %w", err)
	}

	var saves []SaveInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		slotName := file.Name()[:len(file.Name())-5]

		// Get save info
		info, err := sm.getSaveInfo(slotName)
		if err != nil {
			continue
		}

		// Add display name from config
		info.DisplayName = sm.Config.GetDisplayName(slotName)
		info.IsAutoSlot = sm.Config.IsAutoSaveSlot(slotName)
		info.IsQuickSlot = sm.Config.IsQuickSaveSlot(slotName)

		saves = append(saves, info)
	}

	return saves, nil
}

// getSaveInfo extracts metadata from a save file without full loading
func (sm *SaveManager) getSaveInfo(slotName string) (SaveInfo, error) {
	filename := fmt.Sprintf("%s.json", slotName)
	filePath := filepath.Join(sm.SaveDirectory, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return SaveInfo{}, err
	}

	// Only unmarshal the fields we need for listing
	var metadata struct {
		PlayerName  string    `json:"player_name"`
		LocationID  string    `json:"location_id"`
		SaveTime    time.Time `json:"save_time"`
		GameVersion string    `json:"game_version"`
	}

	if err := json.Unmarshal(data, &metadata); err != nil {
		return SaveInfo{}, err
	}

	return SaveInfo{
		SlotName:    slotName,
		PlayerName:  metadata.PlayerName,
		LocationID:  metadata.LocationID,
		SaveTime:    metadata.SaveTime,
		GameVersion: metadata.GameVersion,
	}, nil
}

// DeleteSave deletes a save file
func (sm *SaveManager) DeleteSave(slotName string) error {
	filename := fmt.Sprintf("%s.json", slotName)
	filePath := filepath.Join(sm.SaveDirectory, filename)

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete save file: %w", err)
	}

	fmt.Printf("已删除存档: %s\n", slotName)
	return nil
}

// HasSave checks if a save file exists for the given slot
func (sm *SaveManager) HasSave(slotName string) bool {
	filename := fmt.Sprintf("%s.json", slotName)
	filePath := filepath.Join(sm.SaveDirectory, filename)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetLatestSave returns the most recently modified save
func (sm *SaveManager) GetLatestSave() (SaveInfo, error) {
	saves, err := sm.ListSaves()
	if err != nil {
		return SaveInfo{}, err
	}

	if len(saves) == 0 {
		return SaveInfo{}, fmt.Errorf("no saves found")
	}

	latest := saves[0]
	for _, save := range saves[1:] {
		if save.SaveTime.After(latest.SaveTime) {
			latest = save
		}
	}

	return latest, nil
}

// AutoSave creates an autosave (special slot)
func (sm *SaveManager) AutoSave(
	player *models.Player,
	defeatedEnemies map[string]bool,
	takenItems map[string]bool,
	pendingDrops map[string][]string,
	gameVersion string,
	playerName string,
) error {
	_, err := sm.CreateSave(player, defeatedEnemies, takenItems, pendingDrops, gameVersion, playerName)
	if err != nil {
		return err
	}

	autoSlot := sm.Config.GetAutoSaveSlot()
	return sm.SaveToFile(autoSlot)
}

// QuickSave creates a quicksave (special slot)
func (sm *SaveManager) QuickSave(
	player *models.Player,
	defeatedEnemies map[string]bool,
	takenItems map[string]bool,
	pendingDrops map[string][]string,
	gameVersion string,
	playerName string,
) error {
	_, err := sm.CreateSave(player, defeatedEnemies, takenItems, pendingDrops, gameVersion, playerName)
	if err != nil {
		return err
	}

	quickSlot := sm.Config.GetQuickSaveSlot()
	return sm.SaveToFile(quickSlot)
}

// QuickLoad loads the most recent quicksave
func (sm *SaveManager) QuickLoad() (*SaveData, error) {
	quickSlot := sm.Config.GetQuickSaveSlot()
	return sm.LoadFromFile(quickSlot)
}

// LoadAutoSave loads the autosave
func (sm *SaveManager) LoadAutoSave() (*SaveData, error) {
	autoSlot := sm.Config.GetAutoSaveSlot()
	return sm.LoadFromFile(autoSlot)
}
