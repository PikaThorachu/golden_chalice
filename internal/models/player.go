package models

import (
	"fmt"
)

// Player represents the game character
type Player struct {
	Name               string   `json:"name"`                 // Character name
	Health             int      `json:"health"`               // Current health points
	MaxHealth          int      `json:"max_health"`           // Maximum health points
	CurrentLocationID  string   `json:"current_location_id"`  // Current room/location ID
	Inventory          []string `json:"inventory"`            // Slice of item IDs
	InventorySize      int      `json:"inventory_size"`       // Current inventory capacity
	EquippedWeaponID   *string  `json:"equipped_weapon_id"`   // Currently equipped weapon (nil if none)
	EquippedArmorID    *string  `json:"equipped_armor_id"`    // Currently equipped armor (nil if none)
	EquippedBackpackID *string  `json:"equipped_backpack_id"` // Equipped backpack that increases capacity
	ExperiencePoints   int      `json:"experience_points"`    // For future leveling system
	Level              int      `json:"level"`                // Current level
}

// NewPlayer creates a new player with default starting values
func NewPlayer(name string, startingLocationID string, startingHealth int, startingInventorySize int) *Player {
	return &Player{
		Name:               name,
		Health:             startingHealth,
		MaxHealth:          startingHealth,
		CurrentLocationID:  startingLocationID,
		Inventory:          []string{},
		InventorySize:      startingInventorySize,
		EquippedWeaponID:   nil,
		EquippedArmorID:    nil,
		EquippedBackpackID: nil,
		ExperiencePoints:   0,
		Level:              1,
	}
}

// IsAlive checks if the player is alive
func (p *Player) IsAlive() bool {
	return p.Health > 0
}

// TakeDamage reduces player health and returns actual damage taken
func (p *Player) TakeDamage(amount int) int {
	if amount <= 0 {
		return 0
	}

	actualDamage := amount
	if p.Health-amount < 0 {
		actualDamage = p.Health
		p.Health = 0
	} else {
		p.Health -= amount
	}

	return actualDamage
}

// Heal restores player health and returns actual health restored
func (p *Player) Heal(amount int) int {
	if amount <= 0 {
		return 0
	}

	if !p.IsAlive() {
		return 0
	}

	actualHeal := amount
	if p.Health+amount > p.MaxHealth {
		actualHeal = p.MaxHealth - p.Health
		p.Health = p.MaxHealth
	} else {
		p.Health += amount
	}

	return actualHeal
}

// FullHeal restores player to maximum health
func (p *Player) FullHeal() {
	p.Health = p.MaxHealth
}

// AddItem adds an item to inventory
func (p *Player) AddItem(itemID string) {
	p.Inventory = append(p.Inventory, itemID)
}

// RemoveItem removes an item from inventory
// Returns true if item was found and removed
func (p *Player) RemoveItem(itemID string) bool {
	for i, id := range p.Inventory {
		if id == itemID {
			// Remove by swapping with last element and truncating
			p.Inventory[i] = p.Inventory[len(p.Inventory)-1]
			p.Inventory = p.Inventory[:len(p.Inventory)-1]
			return true
		}
	}
	return false
}

// HasItem checks if player has a specific item in inventory
func (p *Player) HasItem(itemID string) bool {
	for _, id := range p.Inventory {
		if id == itemID {
			return true
		}
	}
	return false
}

// GetItemCount returns how many of a specific item the player has
func (p *Player) GetItemCount(itemID string) int {
	count := 0
	for _, id := range p.Inventory {
		if id == itemID {
			count++
		}
	}
	return count
}

// EquipWeapon equips a weapon if the player has it
// Returns true if successful
func (p *Player) EquipWeapon(itemID string) bool {
	if !p.HasItem(itemID) {
		return false
	}
	p.EquippedWeaponID = &itemID
	return true
}

// EquipArmor equips armor if the player has it
// Returns true if successful
func (p *Player) EquipArmor(itemID string) bool {
	if !p.HasItem(itemID) {
		return false
	}
	p.EquippedArmorID = &itemID
	return true
}

// UnequipWeapon removes the currently equipped weapon
func (p *Player) UnequipWeapon() {
	p.EquippedWeaponID = nil
}

// UnequipArmor removes the currently equipped armor
func (p *Player) UnequipArmor() {
	p.EquippedArmorID = nil
}

// GetEquippedWeaponID returns the equipped weapon ID (empty string if none)
func (p *Player) GetEquippedWeaponID() string {
	if p.EquippedWeaponID == nil {
		return ""
	}
	return *p.EquippedWeaponID
}

// GetEquippedArmorID returns the equipped armor ID (empty string if none)
func (p *Player) GetEquippedArmorID() string {
	if p.EquippedArmorID == nil {
		return ""
	}
	return *p.EquippedArmorID
}

// HasEquippedWeapon checks if player has any weapon equipped
func (p *Player) HasEquippedWeapon() bool {
	return p.EquippedWeaponID != nil
}

// HasEquippedArmor checks if player has any armor equipped
func (p *Player) HasEquippedArmor() bool {
	return p.EquippedArmorID != nil
}

// AddExperience adds experience points and handles leveling up
// Returns true if player leveled up
func (p *Player) AddExperience(amount int) bool {
	if amount <= 0 {
		return false
	}

	p.ExperiencePoints += amount

	// Simple leveling formula: level * 100 XP needed
	neededXP := p.Level * 100
	if p.ExperiencePoints >= neededXP {
		p.LevelUp()
		return true
	}
	return false
}

// LevelUp increases player level and restores health
func (p *Player) LevelUp() {
	p.Level++
	p.MaxHealth += 20
	p.FullHeal()
}

// GetAttackDamage calculates total attack damage including weapon bonus
func (p *Player) GetAttackDamage(baseDamage int, items map[string]Item) int {
	totalDamage := baseDamage

	// Add weapon bonus if equipped
	if p.HasEquippedWeapon() {
		if weapon, exists := items[p.GetEquippedWeaponID()]; exists {
			totalDamage += weapon.GetDamageBonus()
		}
	}

	return totalDamage
}

// GetDefenseValue calculates total defense including armor bonus
func (p *Player) GetDefenseValue(baseDefense int, items map[string]Item) int {
	totalDefense := baseDefense

	// Add armor bonus if equipped
	if p.HasEquippedArmor() {
		if armor, exists := items[p.GetEquippedArmorID()]; exists {
			totalDefense += armor.GetDefenseBonus()
		}
	}

	return totalDefense
}

// MoveTo changes the player's current location
func (p *Player) MoveTo(locationID string) {
	p.CurrentLocationID = locationID
}

// GetCurrentLocation returns the player's current location ID
func (p *Player) GetCurrentLocation() string {
	return p.CurrentLocationID
}

// GetHealthStatus returns a string representation of health percentage
func (p *Player) GetHealthStatus() string {
	percentage := (float64(p.Health) / float64(p.MaxHealth)) * 100
	switch {
	case percentage >= 75:
		return "健康"
	case percentage >= 50:
		return "轻伤"
	case percentage >= 25:
		return "重伤"
	case percentage > 0:
		return "濒死"
	default:
		return "死亡"
	}
}

// GetHealthPercentage returns health as a percentage (0-100)
func (p *Player) GetHealthPercentage() int {
	return int((float64(p.Health) / float64(p.MaxHealth)) * 100)
}

// IsInventoryFull checks if inventory has reached current capacity
func (p *Player) IsInventoryFull() bool {
	return len(p.Inventory) >= p.InventorySize
}

// AddItemWithCheck adds an item to inventory with capacity check
// Returns (success, error message)
func (p *Player) AddItemWithCheck(itemID string) (bool, string) {
	if p.IsInventoryFull() {
		return false, fmt.Sprintf("背包已满 (容量: %d/%d)", len(p.Inventory), p.InventorySize)
	}
	p.Inventory = append(p.Inventory, itemID)
	return true, ""
}

// GetInventoryCapacity returns current and max inventory size
func (p *Player) GetInventoryCapacity() (current, max int) {
	return len(p.Inventory), p.InventorySize
}

// GetInventoryCapacityPercent returns percentage of inventory used
func (p *Player) GetInventoryCapacityPercent() int {
	if p.InventorySize == 0 {
		return 0
	}
	return (len(p.Inventory) * 100) / p.InventorySize
}

// IncreaseInventorySize increases inventory capacity (up to maximum)
// Returns actual amount increased
func (p *Player) IncreaseInventorySize(amount int, maxSize int) int {
	newSize := p.InventorySize + amount
	if newSize > maxSize {
		newSize = maxSize
	}
	increase := newSize - p.InventorySize
	p.InventorySize = newSize
	return increase
}

// DecreaseInventorySize decreases inventory capacity
// Returns actual amount decreased
func (p *Player) DecreaseInventorySize(amount int) int {
	newSize := p.InventorySize - amount
	if newSize < len(p.Inventory) {
		// Can't decrease below current item count
		newSize = len(p.Inventory)
	}
	decrease := p.InventorySize - newSize
	p.InventorySize = newSize
	return decrease
}

// SetInventorySize sets inventory capacity directly (with validation)
func (p *Player) SetInventorySize(newSize int, maxSize int) {
	if newSize > maxSize {
		newSize = maxSize
	}
	if newSize < len(p.Inventory) {
		newSize = len(p.Inventory)
	}
	p.InventorySize = newSize
}

// EquipBackpack equips a backpack and increases inventory capacity
// Returns (success, capacityIncrease, message)
func (p *Player) EquipBackpack(itemID string, sizeBonus int, maxSize int) (bool, int, string) {
	if p.EquippedBackpackID != nil {
		return false, 0, "已经装备了一个背包"
	}

	p.EquippedBackpackID = &itemID
	increase := p.IncreaseInventorySize(sizeBonus, maxSize)
	return true, increase, fmt.Sprintf("装备了背包，背包容量增加了 %d", increase)
}

// UnequipBackpack removes equipped backpack and decreases inventory capacity
// Returns (success, capacityDecrease, message)
func (p *Player) UnequipBackpack() (bool, int, string) {
	if p.EquippedBackpackID == nil {
		return false, 0, "没有装备背包"
	}

	// Store the backpack ID to use for calculating size bonus
	// Note: The caller needs to provide the size bonus since we don't store it here
	p.EquippedBackpackID = nil
	return true, 0, "卸下了背包"
}

// UnequipBackpackWithBonus removes equipped backpack with specific bonus amount
// Returns (success, capacityDecrease, message)
func (p *Player) UnequipBackpackWithBonus(sizeBonus int) (bool, int, string) {
	if p.EquippedBackpackID == nil {
		return false, 0, "没有装备背包"
	}

	// Check if we can decrease without going below current item count
	if p.InventorySize-sizeBonus < len(p.Inventory) {
		return false, 0, fmt.Sprintf("背包中有太多物品，无法卸下背包 (当前: %d/%d)", len(p.Inventory), p.InventorySize)
	}

	p.EquippedBackpackID = nil
	decrease := p.DecreaseInventorySize(sizeBonus)
	return true, decrease, fmt.Sprintf("卸下了背包，背包容量减少了 %d", decrease)
}

// HasBackpackEquipped checks if player has a backpack equipped
func (p *Player) HasBackpackEquipped() bool {
	return p.EquippedBackpackID != nil
}

// GetEquippedBackpackID returns the equipped backpack ID (empty string if none)
func (p *Player) GetEquippedBackpackID() string {
	if p.EquippedBackpackID == nil {
		return ""
	}
	return *p.EquippedBackpackID
}

// ClearInventory removes all items (useful for debugging or new game plus)
func (p *Player) ClearInventory() {
	p.Inventory = []string{}
	p.EquippedWeaponID = nil
	p.EquippedArmorID = nil
	p.EquippedBackpackID = nil
	p.InventorySize = 10 // Reset to default
}

// Reset resets player to starting state (for new game)
func (p *Player) Reset(name string, startingLocationID string, startingHealth int, startingInventorySize int) {
	p.Name = name
	p.Health = startingHealth
	p.MaxHealth = startingHealth
	p.CurrentLocationID = startingLocationID
	p.Inventory = []string{}
	p.InventorySize = startingInventorySize
	p.EquippedWeaponID = nil
	p.EquippedArmorID = nil
	p.EquippedBackpackID = nil
	p.ExperiencePoints = 0
	p.Level = 1
}
