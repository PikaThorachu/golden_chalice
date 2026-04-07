package models

import (
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

// Location represents a place in the game world
type Location struct {
	ID          string   `json:"id"`
	BiomeID     string   `json:"biome_id"`
	Name        Text     `json:"name"`
	Description Text     `json:"description"`
	Exits       []Exit   `json:"exits"`
	EnemyIDs    []string `json:"enemy_ids"`
	ItemIDs     []string `json:"item_ids"`
}

// World contains all locations in the game
type World struct {
	Locations map[string]Location `json:"locations"`
}

// ParseDirection converts user input to a Direction enum
// Supports Chinese format: "往<方向>走" or "往<方向>去"
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

	// Invalid format
	return North, fmt.Errorf("请输入'往<方向>走'或'往<方向>去'的格式")
}

// parseChineseDirection handles "往<方向>走" or "往<方向>去" format
func parseChineseDirection(input string) (Direction, error) {
	// Must be exactly 3 characters: 往 + direction + 走/去
	runes := []rune(input)
	if len(runes) != 3 {
		return North, fmt.Errorf("请输入'往<方向>走'或'往<方向>去'的格式")
	}

	// First character must be '往'
	if runes[0] != '往' {
		return North, fmt.Errorf("请输入'往<方向>走'或'往<方向>去'的格式")
	}

	// Last character must be '走' or '去'
	lastChar := runes[2]
	if lastChar != '走' && lastChar != '去' {
		return North, fmt.Errorf("请输入'往<方向>走'或'往<方向>去'的格式")
	}

	// Middle character is the direction
	dirChar := runes[1]

	// Map Chinese direction characters to Direction enum
	switch dirChar {
	case '北':
		return North, nil
	case '南':
		return South, nil
	case '东':
		return East, nil
	case '西':
		return West, nil
	case '出':
		return Out, nil
	default:
		return North, fmt.Errorf("未知的方向: %c", dirChar)
	}
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
