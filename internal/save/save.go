package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golden_chalice/internal/models"
)

// ============================================================================
// Configuration Structs
// ============================================================================

// SaveConfig represents the configuration for the save system
type SaveConfig struct {
	Enabled            bool              `json:"enabled"`
	AutoSaveEnabled    bool              `json:"auto_save_enabled"`
	AutoSaveInterval   int               `json:"auto_save_interval"`
	MaxSaveSlots       int               `json:"max_save_slots"`
	SaveDirectory      string            `json:"save_directory"`
	CompressSaves      bool              `json:"compress_saves"`
	NamedSlots         []NamedSlotConfig `json:"named_slots"`
	AllowCustomSlots   bool              `json:"allow_custom_slots"`
	MaxCustomSlots     int               `json:"max_custom_slots"`
	SlotNamePattern    string            `json:"slot_name_pattern"`
	AutoGenerateNames  bool              `json:"auto_generate_names"`
	ShowSaveTime       bool              `json:"show_save_time"`
	ShowLocationInList bool              `json:"show_location_in_list"`
}

// NamedSlotConfig represents a predefined save slot
type NamedSlotConfig struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	IsAutoSlot  bool   `json:"is_auto_slot"`
	IsQuickSlot bool   `json:"is_quick_slot"`
	IsReadOnly  bool   `json:"is_read_only"`
	MaxFiles    int    `json:"max_files"`
}

// ============================================================================
// Save Data Structs
// ============================================================================

// SaveData represents the complete save file structure
type SaveData struct {
	Version         string              `json:"version"`
	SaveTime        time.Time           `json:"save_time"`
	PlayerName      string              `json:"player_name"`
	LocationID      string              `json:"location_id"`
	SaveVersion     string              `json:"save_version"`
	Player          *models.Player      `json:"player"`
	DefeatedEnemies map[string]bool     `json:"defeated_enemies"`
	TakenItems      map[string]bool     `json:"taken_items"`
	PendingDrops    map[string][]string `json:"pending_drops"`
	GameVersion     string              `json:"game_version"`
}

// SaveInfo contains metadata about a save file (for listing)
type SaveInfo struct {
	SlotName     string    `json:"slot_name"`
	DisplayName  string    `json:"display_name"`
	PlayerName   string    `json:"player_name"`
	LocationID   string    `json:"location_id"`
	LocationName string    `json:"location_name"`
	SaveTime     time.Time `json:"save_time"`
	GameVersion  string    `json:"game_version"`
	IsAutoSlot   bool      `json:"is_auto_slot"`
	IsQuickSlot  bool      `json:"is_quick_slot"`
}

// ============================================================================
// Save Manager
// ============================================================================

// SaveManager handles save file operations
type SaveManager struct {
	SaveDirectory string
	Config        *SaveConfig
	CurrentSave   *SaveData
	autoSaveTimer *time.Timer
}

// DefaultSaveConfig returns a default save configuration
func DefaultSaveConfig() *SaveConfig {
	return &SaveConfig{
		Enabled:            true,
		AutoSaveEnabled:    true,
		AutoSaveInterval:   5,
		MaxSaveSlots:       10,
		SaveDirectory:      "./saves",
		CompressSaves:      false,
		AllowCustomSlots:   true,
		MaxCustomSlots:     5,
		SlotNamePattern:    "^[a-zA-Z0-9_]+$",
		AutoGenerateNames:  true,
		ShowSaveTime:       true,
		ShowLocationInList: true,
		NamedSlots: []NamedSlotConfig{
			{Name: "autosave", DisplayName: "自动保存", IsAutoSlot: true, IsQuickSlot: false, IsReadOnly: false, MaxFiles: 1},
			{Name: "quicksave", DisplayName: "快速保存", IsAutoSlot: false, IsQuickSlot: true, IsReadOnly: false, MaxFiles: 1},
			{Name: "slot1", DisplayName: "存档 1", IsAutoSlot: false, IsQuickSlot: false, IsReadOnly: false, MaxFiles: 1},
			{Name: "slot2", DisplayName: "存档 2", IsAutoSlot: false, IsQuickSlot: false, IsReadOnly: false, MaxFiles: 1},
			{Name: "slot3", DisplayName: "存档 3", IsAutoSlot: false, IsQuickSlot: false, IsReadOnly: false, MaxFiles: 1},
		},
	}
}

// NewSaveManager creates a new save manager with configuration
func NewSaveManager(saveDir string, configPath string) (*SaveManager, error) {
	config, err := loadSaveConfig(configPath)
	if err != nil {
		return nil, err
	}

	if config.SaveDirectory != "" && config.SaveDirectory != saveDir {
		saveDir = config.SaveDirectory
	}

	if !config.Enabled {
		fmt.Println("Save functionality is disabled in configuration")
	}

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create save directory: %w", err)
	}

	sm := &SaveManager{
		SaveDirectory: saveDir,
		Config:        config,
		CurrentSave:   nil,
	}

	if config.AutoSaveEnabled {
		sm.startAutoSaveTimer()
	}

	return sm, nil
}

// loadSaveConfig loads save configuration from a JSON file
func loadSaveConfig(filePath string) (*SaveConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Save config not found at %s, using defaults\n", filePath)
			return DefaultSaveConfig(), nil
		}
		return nil, fmt.Errorf("failed to read save config file: %w", err)
	}

	var config SaveConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse save config: %w", err)
	}

	if err := validateSaveConfig(&config); err != nil {
		return nil, fmt.Errorf("save config validation failed: %w", err)
	}

	return &config, nil
}

// validateSaveConfig checks if the save configuration is valid
func validateSaveConfig(sc *SaveConfig) error {
	if sc.MaxSaveSlots < 1 {
		return fmt.Errorf("max_save_slots must be at least 1, got %d", sc.MaxSaveSlots)
	}
	if sc.AutoSaveInterval < 1 {
		return fmt.Errorf("auto_save_interval must be at least 1 minute, got %d", sc.AutoSaveInterval)
	}
	if sc.MaxCustomSlots < 0 {
		return fmt.Errorf("max_custom_slots cannot be negative, got %d", sc.MaxCustomSlots)
	}

	names := make(map[string]bool)
	for _, slot := range sc.NamedSlots {
		if names[slot.Name] {
			return fmt.Errorf("duplicate named slot: %s", slot.Name)
		}
		names[slot.Name] = true
	}
	return nil
}

// GetNamedSlot returns a named slot configuration by name
func (sc *SaveConfig) GetNamedSlot(name string) *NamedSlotConfig {
	for i := range sc.NamedSlots {
		if sc.NamedSlots[i].Name == name {
			return &sc.NamedSlots[i]
		}
	}
	return nil
}

// IsValidSlotName checks if a slot name is valid
func (sc *SaveConfig) IsValidSlotName(name string) bool {
	if sc.GetNamedSlot(name) != nil {
		return true
	}
	if sc.AllowCustomSlots && len(name) > 0 && len(name) < 50 {
		return true
	}
	return false
}

// GetDisplayName returns the display name for a slot
func (sc *SaveConfig) GetDisplayName(slotName string) string {
	if slot := sc.GetNamedSlot(slotName); slot != nil {
		return slot.DisplayName
	}
	return slotName
}

// IsAutoSaveSlot checks if a slot is an autosave slot
func (sc *SaveConfig) IsAutoSaveSlot(slotName string) bool {
	if slot := sc.GetNamedSlot(slotName); slot != nil {
		return slot.IsAutoSlot
	}
	return false
}

// IsQuickSaveSlot checks if a slot is a quicksave slot
func (sc *SaveConfig) IsQuickSaveSlot(slotName string) bool {
	if slot := sc.GetNamedSlot(slotName); slot != nil {
		return slot.IsQuickSlot
	}
	return false
}

// GetAutoSaveSlot returns the autosave slot name
func (sc *SaveConfig) GetAutoSaveSlot() string {
	for _, slot := range sc.NamedSlots {
		if slot.IsAutoSlot {
			return slot.Name
		}
	}
	return "autosave"
}

// GetQuickSaveSlot returns the quicksave slot name
func (sc *SaveConfig) GetQuickSaveSlot() string {
	for _, slot := range sc.NamedSlots {
		if slot.IsQuickSlot {
			return slot.Name
		}
	}
	return "quicksave"
}

// startAutoSaveTimer starts the auto-save timer
func (sm *SaveManager) startAutoSaveTimer() {
	if sm.autoSaveTimer != nil {
		sm.autoSaveTimer.Stop()
	}
	duration := time.Duration(sm.Config.AutoSaveInterval) * time.Minute
	sm.autoSaveTimer = time.AfterFunc(duration, func() {
		if sm.CurrentSave != nil && sm.Config.AutoSaveEnabled {
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

// SaveToFile saves the current game state to a file
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

	sm.CurrentSave.SaveTime = time.Now()

	data, err := json.MarshalIndent(sm.CurrentSave, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	filename := fmt.Sprintf("%s.json", slotName)
	filePath := filepath.Join(sm.SaveDirectory, filename)

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

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("save file '%s' not found", slotName)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	var saveData SaveData
	if err := json.Unmarshal(data, &saveData); err != nil {
		return nil, fmt.Errorf("failed to parse save file: %w", err)
	}

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

// ListSaves returns all available save files
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

		info, err := sm.getSaveInfo(slotName)
		if err != nil {
			continue
		}

		info.DisplayName = sm.Config.GetDisplayName(slotName)
		info.IsAutoSlot = sm.Config.IsAutoSaveSlot(slotName)
		info.IsQuickSlot = sm.Config.IsQuickSaveSlot(slotName)

		saves = append(saves, info)
	}

	return saves, nil
}

// getSaveInfo extracts metadata from a save file
func (sm *SaveManager) getSaveInfo(slotName string) (SaveInfo, error) {
	filename := fmt.Sprintf("%s.json", slotName)
	filePath := filepath.Join(sm.SaveDirectory, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return SaveInfo{}, err
	}

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

// HasSave checks if a save file exists
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

// AutoSave creates an autosave
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

// QuickSave creates a quicksave
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
