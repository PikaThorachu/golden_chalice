package models

// Player represents the game character
type Player struct {
	Name              string   `json:"name"`                // Character name
	Health            int      `json:"health"`              // Current health points
	MaxHealth         int      `json:"max_health"`          // Maximum health points
	CurrentLocationID string   `json:"current_location_id"` // Current room/location ID
	Inventory         []string `json:"inventory"`           // Slice of item IDs
	EquippedWeaponID  *string  `json:"equipped_weapon_id"`  // Currently equipped weapon (nil if none)
	EquippedArmorID   *string  `json:"equipped_armor_id"`   // Currently equipped armor (nil if none)
	ExperiencePoints  int      `json:"experience_points"`   // For future leveling system
	Level             int      `json:"level"`               // Current level
}

// NewPlayer creates a new player with default starting values
func NewPlayer(name string, startingLocationID string, startingHealth int) *Player {
	return &Player{
		Name:              name,
		Health:            startingHealth,
		MaxHealth:         startingHealth,
		CurrentLocationID: startingLocationID,
		Inventory:         []string{},
		EquippedWeaponID:  nil,
		EquippedArmorID:   nil,
		ExperiencePoints:  0,
		Level:             1,
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

// IsInventoryFull checks if inventory has reached capacity (optional, for future)
func (p *Player) IsInventoryFull() bool {
	// Default max capacity of 20 items
	return len(p.Inventory) >= 20
}

// GetInventorySize returns current inventory count
func (p *Player) GetInventorySize() int {
	return len(p.Inventory)
}

// ClearInventory removes all items (useful for debugging or new game plus)
func (p *Player) ClearInventory() {
	p.Inventory = []string{}
	p.EquippedWeaponID = nil
	p.EquippedArmorID = nil
}

// Reset resets player to starting state (for new game)
func (p *Player) Reset(name string, startingLocationID string, startingHealth int) {
	p.Name = name
	p.Health = startingHealth
	p.MaxHealth = startingHealth
	p.CurrentLocationID = startingLocationID
	p.Inventory = []string{}
	p.EquippedWeaponID = nil
	p.EquippedArmorID = nil
	p.ExperiencePoints = 0
	p.Level = 1
}
