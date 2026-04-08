package game

import (
	"fmt"

	"golden_chalice/internal/models"
)

// PlayerController handles player actions and state management
type PlayerController struct {
	Player *models.Player
	Config *models.Config
	World  *models.World
	Items  map[string]models.Item
}

// NewPlayerController creates a new player controller
func NewPlayerController(config *models.Config, world *models.World, items map[string]models.Item) *PlayerController {
	return &PlayerController{
		Player: nil, // Will be created when starting a new game
		Config: config,
		World:  world,
		Items:  items,
	}
}

// NewGame creates a new player character
func (pc *PlayerController) NewGame(playerName string) {
	pc.Player = models.NewPlayer(
		playerName,
		pc.Config.GetStartingLocationID(),
		pc.Config.GetStartingHealth(),
	)
}

// GetPlayer returns the current player instance
func (pc *PlayerController) GetPlayer() *models.Player {
	return pc.Player
}

// Move attempts to move the player in a direction
// Returns error message if movement fails
func (pc *PlayerController) Move(direction models.Direction) error {
	if pc.Player == nil {
		return fmt.Errorf("no active game")
	}

	// Get destination using world logic
	destID, err := pc.World.GetDestination(
		pc.Player.GetCurrentLocation(),
		direction,
		pc.Player.Inventory,
	)
	if err != nil {
		return err
	}

	// Check for enemies in destination (optional - could be checked after move)
	destLocation, _ := pc.World.GetLocation(destID)
	if len(destLocation.EnemyIDs) > 0 {
		// Enemy present - you might want to handle this differently
		// For now, allow movement but trigger combat in game loop
		fmt.Println("⚠️ 前方有敌人！")
	}

	// Move player
	pc.Player.MoveTo(destID)
	return nil
}

// TakeItem attempts to pick up an item from the current location
// Returns error message if taking fails
func (pc *PlayerController) TakeItem(itemID string) error {
	if pc.Player == nil {
		return fmt.Errorf("no active game")
	}

	// Get current location
	location, err := pc.World.GetLocation(pc.Player.GetCurrentLocation())
	if err != nil {
		return err
	}

	// Check if item exists at location
	itemExists := false
	for _, id := range location.ItemIDs {
		if id == itemID {
			itemExists = true
			break
		}
	}

	if !itemExists {
		return fmt.Errorf("这里没有 %s", itemID)
	}

	// Add to inventory
	pc.Player.AddItem(itemID)

	// Note: In a real implementation, you'd need to update the world state
	// to remove the item from the location. This requires a GameState that
	// tracks taken items separately from the static world data.

	return nil
}

// UseItem attempts to use an item from inventory
func (pc *PlayerController) UseItem(itemID string) (string, error) {
	if pc.Player == nil {
		return "", fmt.Errorf("no active game")
	}

	if !pc.Player.HasItem(itemID) {
		return "", fmt.Errorf("你没有 %s", itemID)
	}

	item, exists := pc.Items[itemID]
	if !exists {
		return "", fmt.Errorf("物品 %s 不存在", itemID)
	}

	if !item.IsUsable() {
		return "", fmt.Errorf("不能使用 %s", item.GetDisplayName(pc.Config))
	}

	if item.IsConsumable() {
		// Use consumable
		healthRestored, manaRestored, message := item.Use()

		if healthRestored > 0 {
			pc.Player.Heal(healthRestored)
		}
		// Handle mana restoration when implemented
		_ = manaRestored

		// Remove consumed item
		pc.Player.RemoveItem(itemID)

		return message, nil
	}

	if item.IsEquippable() {
		// Equip weapon or armor
		if item.IsWeapon() {
			pc.Player.EquipWeapon(itemID)
			return fmt.Sprintf("装备了 %s", item.GetDisplayName(pc.Config)), nil
		}
		if item.IsArmor() {
			pc.Player.EquipArmor(itemID)
			return fmt.Sprintf("装备了 %s", item.GetDisplayName(pc.Config)), nil
		}
	}

	return "什么也没发生", nil
}

// DropItem removes an item from inventory and would place it in the world
// (Full implementation would need to update world state)
func (pc *PlayerController) DropItem(itemID string) error {
	if pc.Player == nil {
		return fmt.Errorf("no active game")
	}

	if !pc.Player.HasItem(itemID) {
		return fmt.Errorf("你没有 %s", itemID)
	}

	// Unequip if it was equipped
	if pc.Player.GetEquippedWeaponID() == itemID {
		pc.Player.UnequipWeapon()
	}
	if pc.Player.GetEquippedArmorID() == itemID {
		pc.Player.UnequipArmor()
	}

	pc.Player.RemoveItem(itemID)
	return nil
}

// ShowInventory displays the player's inventory
func (pc *PlayerController) ShowInventory() {
	if pc.Player == nil {
		fmt.Println("没有活跃的游戏")
		return
	}

	if pc.Player.GetInventorySize() == 0 {
		fmt.Println("背包是空的。")
		return
	}

	fmt.Println("\n=== 背包 ===")
	for _, itemID := range pc.Player.Inventory {
		if item, exists := pc.Items[itemID]; exists {
			equipped := ""
			if pc.Player.GetEquippedWeaponID() == itemID {
				equipped = " [已装备]"
			}
			if pc.Player.GetEquippedArmorID() == itemID {
				equipped = " [已装备]"
			}
			fmt.Printf("  - %s%s\n", pc.Config.GetDisplayText(item.Name), equipped)
		}
	}
	fmt.Printf("\n共 %d 件物品\n", pc.Player.GetInventorySize())
}

// ShowStatus displays player status (health, location, equipment)
func (pc *PlayerController) ShowStatus() {
	if pc.Player == nil {
		fmt.Println("没有活跃的游戏")
		return
	}

	fmt.Println("\n=== 角色状态 ===")
	fmt.Printf("名称: %s\n", pc.Player.Name)
	fmt.Printf("等级: %d\n", pc.Player.Level)
	fmt.Printf("生命: %d/%d (%s)\n", pc.Player.Health, pc.Player.MaxHealth, pc.Player.GetHealthStatus())
	fmt.Printf("经验: %d / %d\n", pc.Player.ExperiencePoints, pc.Player.Level*100)

	// Show equipped items
	fmt.Println("\n装备:")
	if pc.Player.HasEquippedWeapon() {
		if weapon, exists := pc.Items[pc.Player.GetEquippedWeaponID()]; exists {
			fmt.Printf("  武器: %s (+%d攻击)\n",
				pc.Config.GetDisplayText(weapon.Name),
				weapon.GetDamageBonus())
		}
	} else {
		fmt.Println("  武器: 无")
	}

	if pc.Player.HasEquippedArmor() {
		if armor, exists := pc.Items[pc.Player.GetEquippedArmorID()]; exists {
			fmt.Printf("  护甲: %s (+%d防御)\n",
				pc.Config.GetDisplayText(armor.Name),
				armor.GetDefenseBonus())
		}
	} else {
		fmt.Println("  护甲: 无")
	}

	// Show location
	location, err := pc.World.GetLocation(pc.Player.GetCurrentLocation())
	if err == nil {
		fmt.Printf("\n当前位置: %s\n", pc.Config.GetDisplayText(location.Name))
	}
}

// HealPlayer heals the player by a certain amount
func (pc *PlayerController) HealPlayer(amount int) int {
	if pc.Player == nil {
		return 0
	}
	return pc.Player.Heal(amount)
}

// DamagePlayer damages the player and returns actual damage
func (pc *PlayerController) DamagePlayer(amount int) int {
	if pc.Player == nil {
		return 0
	}
	return pc.Player.TakeDamage(amount)
}

// IsPlayerAlive checks if the player is alive
func (pc *PlayerController) IsPlayerAlive() bool {
	return pc.Player != nil && pc.Player.IsAlive()
}

// GetCurrentLocation returns the current location ID
func (pc *PlayerController) GetCurrentLocation() string {
	if pc.Player == nil {
		return ""
	}
	return pc.Player.GetCurrentLocation()
}
