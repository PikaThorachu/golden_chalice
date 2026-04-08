package save

import (
	"encoding/json"
	"fmt"
	"os"
)

// SaveConfig represents the configuration for the save system
type SaveConfig struct {
	// General settings
	Enabled          bool   `json:"enabled"`            // Master toggle for save functionality
	AutoSaveEnabled  bool   `json:"auto_save_enabled"`  // Enable/disable autosave
	AutoSaveInterval int    `json:"auto_save_interval"` // Minutes between autosaves
	MaxSaveSlots     int    `json:"max_save_slots"`     // Maximum number of named save slots
	SaveDirectory    string `json:"save_directory"`     // Directory for save files
	CompressSaves    bool   `json:"compress_saves"`     // Compress save files (future feature)

	// Named slots configuration
	NamedSlots       []NamedSlotConfig `json:"named_slots"`        // Predefined named slots
	AllowCustomSlots bool              `json:"allow_custom_slots"` // Allow custom slot names
	MaxCustomSlots   int               `json:"max_custom_slots"`   // Maximum custom slots if allowed

	// Slot naming
	SlotNamePattern   string `json:"slot_name_pattern"`   // Regex pattern for valid slot names
	AutoGenerateNames bool   `json:"auto_generate_names"` // Auto-generate names for unnamed saves

	// Display settings
	ShowSaveTime       bool `json:"show_save_time"`        // Show save time in listings
	ShowLocationInList bool `json:"show_location_in_list"` // Show location in save listings
}

// NamedSlotConfig represents a predefined save slot
type NamedSlotConfig struct {
	Name        string `json:"name"`          // Slot name (e.g., "slot1", "autosave")
	DisplayName string `json:"display_name"`  // Human-readable name (e.g., "存档 1", "自动保存")
	IsAutoSlot  bool   `json:"is_auto_slot"`  // Is this an autosave slot?
	IsQuickSlot bool   `json:"is_quick_slot"` // Is this a quicksave slot?
	IsReadOnly  bool   `json:"is_read_only"`  // Can player overwrite this slot?
	MaxFiles    int    `json:"max_files"`     // For rotating saves (e.g., keep 3 autosaves)
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
			{
				Name:        "autosave",
				DisplayName: "自动保存",
				IsAutoSlot:  true,
				IsQuickSlot: false,
				IsReadOnly:  false,
				MaxFiles:    1,
			},
			{
				Name:        "quicksave",
				DisplayName: "快速保存",
				IsAutoSlot:  false,
				IsQuickSlot: true,
				IsReadOnly:  false,
				MaxFiles:    1,
			},
			{
				Name:        "slot1",
				DisplayName: "存档 1",
				IsAutoSlot:  false,
				IsQuickSlot: false,
				IsReadOnly:  false,
				MaxFiles:    1,
			},
			{
				Name:        "slot2",
				DisplayName: "存档 2",
				IsAutoSlot:  false,
				IsQuickSlot: false,
				IsReadOnly:  false,
				MaxFiles:    1,
			},
			{
				Name:        "slot3",
				DisplayName: "存档 3",
				IsAutoSlot:  false,
				IsQuickSlot: false,
				IsReadOnly:  false,
				MaxFiles:    1,
			},
		},
	}
}

// LoadSaveConfig loads save configuration from a JSON file
func LoadSaveConfig(filePath string) (*SaveConfig, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		// If file doesn't exist, return default config
		if os.IsNotExist(err) {
			fmt.Printf("Save config not found at %s, using defaults\n", filePath)
			return DefaultSaveConfig(), nil
		}
		return nil, fmt.Errorf("failed to read save config file: %w", err)
	}

	// Unmarshal JSON
	var config SaveConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse save config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("save config validation failed: %w", err)
	}

	return &config, nil
}

// Validate checks if the save configuration is valid
func (sc *SaveConfig) Validate() error {
	if sc.MaxSaveSlots < 1 {
		return fmt.Errorf("max_save_slots must be at least 1, got %d", sc.MaxSaveSlots)
	}

	if sc.AutoSaveInterval < 1 {
		return fmt.Errorf("auto_save_interval must be at least 1 minute, got %d", sc.AutoSaveInterval)
	}

	if sc.MaxCustomSlots < 0 {
		return fmt.Errorf("max_custom_slots cannot be negative, got %d", sc.MaxCustomSlots)
	}

	// Validate named slots have unique names
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
	// Check if it's a named slot
	if sc.GetNamedSlot(name) != nil {
		return true
	}

	// Check custom slots if allowed
	if sc.AllowCustomSlots {
		// Simple length check (in production, use regex pattern)
		if len(name) > 0 && len(name) < 50 {
			return true
		}
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
