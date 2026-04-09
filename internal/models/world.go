package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Direction represents cardinal and relative directions for movement
type Direction int

const (
	North Direction = iota
	South
	East
	West
	Northwest
	Northeast
	Southwest
	Southeast
	Up
	Down
	Out
	In
)

// String returns the English string representation of a Direction
func (d Direction) String() string {
	directions := [...]string{
		"north",
		"south",
		"east",
		"west",
		"northwest",
		"northeast",
		"southwest",
		"southeast",
		"up",
		"down",
		"out",
		"in",
	}

	if int(d) < 0 || int(d) >= len(directions) {
		return "unknown"
	}
	return directions[d]
}

// MarshalJSON implements the json.Marshaler interface
func (d Direction) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *Direction) UnmarshalJSON(data []byte) error {
	var dirString string
	if err := json.Unmarshal(data, &dirString); err != nil {
		return err
	}

	switch dirString {
	case "north":
		*d = North
	case "south":
		*d = South
	case "east":
		*d = East
	case "west":
		*d = West
	case "northwest":
		*d = Northwest
	case "northeast":
		*d = Northeast
	case "southwest":
		*d = Southwest
	case "southeast":
		*d = Southeast
	case "up":
		*d = Up
	case "down":
		*d = Down
	case "out":
		*d = Out
	case "in":
		*d = In
	default:
		return fmt.Errorf("invalid direction: %s", dirString)
	}
	return nil
}

// Text holds multilingual text for display
type Text struct {
	Chinese string `json:"chinese"`
	Pinyin  string `json:"pinyin"`
	English string `json:"english"`
}

// Exit represents a connection from one location to another
type Exit struct {
	Direction     Direction `json:"direction"`
	DestinationID string    `json:"destination"`
	RequiredItem  *string   `json:"requires_item_id"` // nil means no item required
}

// Room represents a larger area containing multiple locations
type Room struct {
	ID          string `json:"id"`
	Name        Text   `json:"name"`
	Description Text   `json:"description"`
}

// Location represents a place within a Room
type Location struct {
	ID       string   `json:"id"`
	RoomID   *string  `json:"room_id"` // Optional reference to a Room (nil if no room)
	BiomeID  string   `json:"biome_id"`
	Name     Text     `json:"name"`
	Exits    []Exit   `json:"exits"`
	EnemyIDs []string `json:"enemy_ids"`
	ItemIDs  []string `json:"item_ids"`
}

// World contains all locations and rooms
type World struct {
	Locations map[string]Location `json:"locations"`
	Rooms     map[string]Room     `json:"rooms"`
}

// ParseDirection converts user input to a Direction enum
// Supports Chinese format: "往<方向>走" or "往<方向>去" (方向 can be 1-2 characters)
// Supports English format: "Go <direction>" or "Walk <direction>"
func ParseDirection(input string) (Direction, error) {
	// Trim whitespace
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return North, fmt.Errorf("方向不能为空")
	}

	// Check for Chinese format (starts with '往')
	if strings.HasPrefix(trimmed, "往") {
		return parseChineseDirection(trimmed)
	}

	// Check for English format (case insensitive)
	lowerInput := strings.ToLower(trimmed)
	if strings.HasPrefix(lowerInput, "go ") || strings.HasPrefix(lowerInput, "walk ") {
		return parseEnglishDirection(lowerInput)
	}

	// Also try to parse as raw Chinese direction (without 往 prefix)
	// This handles inputs like "北", "西北", etc.
	dir, err := parseRawChineseDirection(trimmed)
	if err == nil {
		return dir, nil
	}

	// Invalid format
	return North, fmt.Errorf("请输入'往<方向>走'或'往<方向>去'的格式")
}

// parseRawChineseDirection handles Chinese direction without prefix
func parseRawChineseDirection(input string) (Direction, error) {
	// Remove any trailing 走 or 去
	cleaned := strings.TrimSuffix(input, "走")
	cleaned = strings.TrimSuffix(cleaned, "去")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return North, fmt.Errorf("请输入方向")
	}

	// Handle single-character directions
	switch cleaned {
	case "北":
		return North, nil
	case "南":
		return South, nil
	case "东":
		return East, nil
	case "西":
		return West, nil
	case "出":
		return Out, nil
	}

	// Handle two-character diagonal directions
	switch cleaned {
	case "西北":
		return Northwest, nil
	case "东北":
		return Northeast, nil
	case "西南":
		return Southwest, nil
	case "东南":
		return Southeast, nil
	}

	return North, fmt.Errorf("未知的方向: %s", cleaned)
}

// parseChineseDirection handles Chinese direction formats
// Supports: 北, 南, 东, 西, 西北, 东北, 西南, 东南, 出
// With optional prefix: 往北走, 往北去, 往西北走, 往西北去
func parseChineseDirection(input string) (Direction, error) {
	// Remove prefix "往" and suffixes "走"/"去" in any combination
	// Using TrimPrefix/TrimSuffix is safe - they return original string if not found
	trimmed := strings.TrimPrefix(input, "往")
	trimmed = strings.TrimSuffix(trimmed, "走")
	trimmed = strings.TrimSuffix(trimmed, "去")

	// Trim any remaining spaces
	trimmed = strings.TrimSpace(trimmed)

	if trimmed == "" {
		return North, fmt.Errorf("请输入方向")
	}

	// Handle single-character directions
	switch trimmed {
	case "北":
		return North, nil
	case "南":
		return South, nil
	case "东":
		return East, nil
	case "西":
		return West, nil
	case "出":
		return Out, nil
	}

	// Handle two-character diagonal directions
	switch trimmed {
	case "西北":
		return Northwest, nil
	case "东北":
		return Northeast, nil
	case "西南":
		return Southwest, nil
	case "东南":
		return Southeast, nil
	}

	return North, fmt.Errorf("未知的方向: %s", trimmed)
}

// parseEnglishDirection handles "Go <direction>" or "Walk <direction>" format
func parseEnglishDirection(input string) (Direction, error) {
	// Split into words
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return North, fmt.Errorf("Please use 'Go <direction>' or 'Walk <direction>'")
	}

	// First word must be "go" or "walk"
	verb := parts[0]
	if verb != "go" && verb != "walk" {
		return North, fmt.Errorf("Please use 'Go <direction>' or 'Walk <direction>'")
	}

	// Second word is the direction
	dirWord := parts[1]

	// Map English direction words to Direction enum
	switch dirWord {
	case "north":
		return North, nil
	case "south":
		return South, nil
	case "east":
		return East, nil
	case "west":
		return West, nil
	case "northwest":
		return Northwest, nil
	case "northeast":
		return Northeast, nil
	case "southwest":
		return Southwest, nil
	case "southeast":
		return Southeast, nil
	case "out":
		return Out, nil
	default:
		return North, fmt.Errorf("未知的方向: %s", dirWord)
	}
}

// GetLocation retrieves a location by ID
func (w *World) GetLocation(id string) (Location, error) {
	location, exists := w.Locations[id]
	if !exists {
		return Location{}, fmt.Errorf("location '%s' not found", id)
	}
	return location, nil
}

// GetRoom retrieves a room by ID
func (w *World) GetRoom(id string) (Room, error) {
	room, exists := w.Rooms[id]
	if !exists {
		return Room{}, fmt.Errorf("room '%s' not found", id)
	}
	return room, nil
}

// GetCurrentRoom returns the room the player is currently in (if any)
func (w *World) GetCurrentRoom(locationID string) (*Room, error) {
	location, err := w.GetLocation(locationID)
	if err != nil {
		return nil, err
	}

	if location.RoomID == nil {
		return nil, nil // No room associated
	}

	room, err := w.GetRoom(*location.RoomID)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// GetExit finds an exit in the specified location by direction
func (w *World) GetExit(locationID string, dir Direction) (Exit, error) {
	location, err := w.GetLocation(locationID)
	if err != nil {
		return Exit{}, err
	}

	for _, exit := range location.Exits {
		if exit.Direction == dir {
			return exit, nil
		}
	}

	// No exit in that direction - immersive error message
	return Exit{}, fmt.Errorf("一堵石墙挡住了你的去路")
}

// CanUseExit checks if the player can use the given exit
// Returns (canUse, missingItemID)
func (w *World) CanUseExit(playerInventory []string, exit Exit) (bool, string) {
	// If no item required, exit is usable
	if exit.RequiredItem == nil {
		return true, ""
	}

	// Check if player has the required item
	requiredItem := *exit.RequiredItem
	for _, item := range playerInventory {
		if item == requiredItem {
			return true, ""
		}
	}

	// Missing required item
	return false, requiredItem
}

// GetDestination validates movement and returns the destination location ID
func (w *World) GetDestination(locationID string, dir Direction, playerInventory []string) (string, error) {
	// Get the exit
	exit, err := w.GetExit(locationID, dir)
	if err != nil {
		return "", err
	}

	// Check if player can use the exit
	canUse, missingItem := w.CanUseExit(playerInventory, exit)
	if !canUse {
		return "", fmt.Errorf("你需要 %s 才能通过这里", missingItem)
	}

	return exit.DestinationID, nil
}

// GetAvailableExits returns all directions the player can use from a location
func (w *World) GetAvailableExits(locationID string, playerInventory []string) []Direction {
	location, err := w.GetLocation(locationID)
	if err != nil {
		return []Direction{}
	}

	var available []Direction
	for _, exit := range location.Exits {
		canUse, _ := w.CanUseExit(playerInventory, exit)
		if canUse {
			available = append(available, exit.Direction)
		}
	}

	return available
}

// IsExitLocked checks if an exit requires an item (without checking inventory)
func (w *World) IsExitLocked(locationID string, dir Direction) (bool, string) {
	exit, err := w.GetExit(locationID, dir)
	if err != nil {
		return false, ""
	}

	if exit.RequiredItem != nil {
		return true, *exit.RequiredItem
	}
	return false, ""
}

// GetExitDescription returns a user-friendly description of an exit
func (w *World) GetExitDescription(locationID string, dir Direction) string {
	exit, err := w.GetExit(locationID, dir)
	if err != nil {
		return ""
	}

	if exit.RequiredItem != nil {
		return fmt.Sprintf("一扇锁着的门 (需要 %s)", *exit.RequiredItem)
	}
	return "一条畅通的道路"
}
