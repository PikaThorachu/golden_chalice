package game

import (
	"fmt"
	"slices"
	"strings"

	"golden_chalice/internal/models"
)

// GameState represents the entire state of the game at runtime
type GameState struct {
	// Core game data (loaded from JSON)
	Config  *models.Config
	World   *models.World
	Items   map[string]models.Item
	Enemies map[string]models.Enemy
	Biomes  map[string]models.Biome

	// Dynamic game state
	Player          *models.Player
	DefeatedEnemies map[string]bool     // locationID -> true (enemy defeated)
	TakenItems      map[string]bool     // locationID -> true (item taken)
	PendingDrops    map[string][]string // locationID -> []itemID (items on ground)
	GameOver        bool
	GameWon         bool
}

// NewGameState creates a new game state with loaded static data
func NewGameState(
	config *models.Config,
	world *models.World,
	items map[string]models.Item,
	enemies map[string]models.Enemy,
	biomes map[string]models.Biome,
) *GameState {
	return &GameState{
		Config:          config,
		World:           world,
		Items:           items,
		Enemies:         enemies,
		Biomes:          biomes,
		DefeatedEnemies: make(map[string]bool),
		TakenItems:      make(map[string]bool),
		PendingDrops:    make(map[string][]string),
		GameOver:        false,
		GameWon:         false,
	}
}

// NewGame creates a new player and starts a fresh game
func (gs *GameState) NewGame(playerName string) {
	gs.Player = models.NewPlayer(
		playerName,
		gs.Config.GetStartingLocationID(),
		gs.Config.GetStartingHealth(),
	)
	gs.DefeatedEnemies = make(map[string]bool)
	gs.TakenItems = make(map[string]bool)
	gs.PendingDrops = make(map[string][]string)
	gs.GameOver = false
	gs.GameWon = false
}

// GetCurrentRoom returns the current location the player is in
func (gs *GameState) GetCurrentRoom() (models.Location, error) {
	return gs.World.GetLocation(gs.Player.CurrentLocationID)
}

// GetCurrentBiome returns the biome of the current location
func (gs *GameState) GetCurrentBiome() (models.Biome, error) {
	room, err := gs.GetCurrentRoom()
	if err != nil {
		return models.Biome{}, err
	}
	return gs.Biomes[room.BiomeID], nil
}

// HasEnemyAtCurrentLocation checks if there's an undefeated enemy at current location
func (gs *GameState) HasEnemyAtCurrentLocation() bool {
	room, err := gs.GetCurrentRoom()
	if err != nil {
		return false
	}

	// Check if enemy already defeated
	if gs.DefeatedEnemies[room.ID] {
		return false
	}

	return len(room.EnemyIDs) > 0
}

// GetCurrentEnemy returns the first undefeated enemy at current location
func (gs *GameState) GetCurrentEnemy() (*models.Enemy, error) {
	room, err := gs.GetCurrentRoom()
	if err != nil {
		return nil, err
	}

	// Check if enemy already defeated
	if gs.DefeatedEnemies[room.ID] {
		return nil, nil
	}

	if len(room.EnemyIDs) == 0 {
		return nil, nil
	}

	enemy, exists := gs.Enemies[room.EnemyIDs[0]]
	if !exists {
		return nil, fmt.Errorf("enemy '%s' not found", room.EnemyIDs[0])
	}

	return &enemy, nil
}

// DefeatEnemy marks the enemy at current location as defeated
func (gs *GameState) DefeatEnemy() {
	room, err := gs.GetCurrentRoom()
	if err != nil {
		return
	}
	gs.DefeatedEnemies[room.ID] = true
}

// GetItemsAtCurrentLocation returns items that haven't been taken yet
func (gs *GameState) GetItemsAtCurrentLocation() []string {
	room, err := gs.GetCurrentRoom()
	if err != nil {
		return []string{}
	}

	// If items already taken, return empty
	if gs.TakenItems[room.ID] {
		return []string{}
	}

	// Also check for pending drops (items dropped by enemies)
	pendingItems := gs.PendingDrops[room.ID]

	// Combine with static items
	allItems := append([]string{}, room.ItemIDs...)
	allItems = append(allItems, pendingItems...)

	return allItems
}

// TakeItem removes an item from the current location and adds to player inventory
func (gs *GameState) TakeItem(itemID string) error {
	room, err := gs.GetCurrentRoom()
	if err != nil {
		return err
	}

	// Check if static items already taken (optimization)
	if gs.TakenItems[room.ID] && len(room.ItemIDs) > 0 {
		// Static items are gone, but check pending drops still
		pendingItems := gs.PendingDrops[room.ID]
		if len(pendingItems) == 0 {
			return fmt.Errorf("这里已经没有物品了")
		}
	}

	// Check static items
	if slices.Contains(room.ItemIDs, itemID) {
		// Don't allow taking if already taken
		if gs.TakenItems[room.ID] {
			return fmt.Errorf("这个物品已经被拿走了")
		}

		gs.Player.AddItem(itemID)
		// Mark all static items as taken (simplified - all or nothing)
		// For multiple items, you'd want per-item tracking
		gs.TakenItems[room.ID] = true
		return nil
	}

	// Check pending drops (items dropped by enemies)
	pendingItems := gs.PendingDrops[room.ID]
	for i, id := range pendingItems {
		if id == itemID {
			gs.Player.AddItem(itemID)
			// Remove from pending drops
			gs.PendingDrops[room.ID] = append(pendingItems[:i], pendingItems[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("这里没有 %s", itemID)
}

// AddPendingDrop adds an item to the ground at a location
func (gs *GameState) AddPendingDrop(locationID string, itemID string) {
	gs.PendingDrops[locationID] = append(gs.PendingDrops[locationID], itemID)
}

// Move attempts to move the player in a direction
func (gs *GameState) Move(direction models.Direction) error {
	// Check if there's an enemy blocking movement
	if gs.HasEnemyAtCurrentLocation() {
		enemy, _ := gs.GetCurrentEnemy()
		return fmt.Errorf("有 %s 挡住了去路！", gs.GetDisplayName(enemy.Name))
	}

	// Get destination using world logic
	destID, err := gs.World.GetDestination(
		gs.Player.CurrentLocationID,
		direction,
		gs.Player.Inventory,
	)
	if err != nil {
		return err
	}

	// Move player
	gs.Player.CurrentLocationID = destID
	return nil
}

// GetAvailableExits returns exits the player can use from current location
func (gs *GameState) GetAvailableExits() []models.Direction {
	return gs.World.GetAvailableExits(gs.Player.CurrentLocationID, gs.Player.Inventory)
}

// CheckWinCondition checks if player has won the game
func (gs *GameState) CheckWinCondition() bool {
	if gs.GameWon {
		return true
	}

	winItemID := gs.Config.GetWinConditionItemID()
	if gs.Player.HasItem(winItemID) {
		gs.GameWon = true
		gs.GameOver = true
		return true
	}
	return false
}

// CheckLossCondition checks if player has lost the game
func (gs *GameState) CheckLossCondition() bool {
	if !gs.Player.IsAlive() {
		gs.GameOver = true
		return true
	}
	return false
}

// ProcessCombatTurn handles one turn of combat between player and enemy
// Returns (playerVictory, playerDefeated, message)
func (gs *GameState) ProcessCombatTurn(enemy *models.Enemy, playerAttack bool) (bool, bool, string) {
	if playerAttack {
		// Player attacks
		baseDamage := gs.Config.CombatSettings.BaseDamageMin
		damage := gs.Player.GetAttackDamage(baseDamage, gs.Items)
		actualDamage := enemy.TakeDamage(damage)

		if !enemy.IsAlive() {
			// Enemy defeated
			exp := enemy.GetExperienceReward()
			gs.Player.AddExperience(exp)

			// Handle drops
			drops := enemy.GetDropItems()
			for _, dropID := range drops {
				gs.AddPendingDrop(gs.Player.CurrentLocationID, dropID)
			}

			gs.DefeatEnemy()
			return true, false, fmt.Sprintf("你对 %s 造成了 %d 点伤害！击败了敌人！获得 %d 点经验值。",
				gs.GetDisplayName(enemy.Name), actualDamage, exp)
		}

		return false, false, fmt.Sprintf("你对 %s 造成了 %d 点伤害！", gs.GetDisplayName(enemy.Name), actualDamage)
	} else {
		// Enemy attacks
		damage := enemy.CalculateDamage()
		actualDamage := gs.Player.TakeDamage(damage)

		if !gs.Player.IsAlive() {
			return false, true, fmt.Sprintf("%s 对你造成了 %d 点伤害！你被击败了...",
				gs.GetDisplayName(enemy.Name), actualDamage)
		}

		return false, false, fmt.Sprintf("%s 对你造成了 %d 点伤害！", gs.GetDisplayName(enemy.Name), actualDamage)
	}
}

// AttemptFlee attempts to flee from combat
// Returns (success, message)
func (gs *GameState) AttemptFlee() (bool, string) {
	fleeRate := gs.Config.CombatSettings.FleeSuccessRate

	// Simple random check (in real implementation, use rand.Intn(100))
	// For now, assume always successful if rate is 100
	if fleeRate >= 100 {
		return true, "你成功逃跑了！"
	}

	// Placeholder for actual flee logic
	return true, "你成功逃跑了！"
}

// UseItem uses an item from player inventory
func (gs *GameState) UseItem(itemID string) (string, error) {
	if !gs.Player.HasItem(itemID) {
		return "", fmt.Errorf("你没有这个物品")
	}

	item, exists := gs.Items[itemID]
	if !exists {
		return "", fmt.Errorf("物品不存在")
	}

	if !item.IsUsable() {
		return "", fmt.Errorf("不能使用这个物品")
	}

	if item.IsConsumable() {
		healthRestored, _, message := item.Use()
		if healthRestored > 0 {
			gs.Player.Heal(healthRestored)
		}
		gs.Player.RemoveItem(itemID)
		return message, nil
	}

	if item.IsEquippable() {
		if item.IsWeapon() {
			gs.Player.EquipWeapon(itemID)
			return fmt.Sprintf("装备了 %s", gs.GetDisplayName(item.Name)), nil
		}
		if item.IsArmor() {
			gs.Player.EquipArmor(itemID)
			return fmt.Sprintf("装备了 %s", gs.GetDisplayName(item.Name)), nil
		}
	}

	return "什么也没发生", nil
}

// GetDisplayName returns formatted name based on config preferences
func (gs *GameState) GetDisplayName(text models.Text) string {
	return gs.Config.GetDisplayText(text)
}

// GetDisplayDescription returns formatted description based on config preferences
func (gs *GameState) GetDisplayDescription(text models.Text) string {
	return gs.Config.GetDisplayText(text)
}

// GetCurrentRoomDescription returns formatted current room description
func (gs *GameState) GetCurrentRoomDescription() string {
	room, err := gs.GetCurrentRoom()
	if err != nil {
		return "无法获取房间描述"
	}

	var result strings.Builder

	// Room name
	result.WriteString("\n=== ")
	result.WriteString(gs.GetDisplayName(room.Name))
	result.WriteString(" ===\n\n")

	// Room description
	result.WriteString(gs.GetDisplayDescription(room.Description))
	result.WriteString("\n\n")

	// Biome ambient description
	biome, err := gs.GetCurrentBiome()
	if err == nil && (biome.AmbientDescription.Chinese != "" || biome.AmbientDescription.English != "") {
		result.WriteString(gs.GetDisplayDescription(biome.AmbientDescription))
		result.WriteString("\n\n")
	}

	// Items in room
	items := gs.GetItemsAtCurrentLocation()
	if len(items) > 0 {
		result.WriteString("你可以看到: ")
		for i, itemID := range items {
			if i > 0 {
				result.WriteString(", ")
			}
			if item, exists := gs.Items[itemID]; exists {
				result.WriteString(gs.GetDisplayName(item.Name))
			}
		}
		result.WriteString("\n\n")
	}

	// Available exits
	exits := gs.GetAvailableExits()
	if len(exits) > 0 {
		result.WriteString("出口: ")
		exitStrings := make([]string, len(exits))
		for i, exit := range exits {
			exitStrings[i] = exit.String()
		}
		result.WriteString(strings.Join(exitStrings, ", "))
		result.WriteString("\n")
	}

	return result.String()
}

// GetPlayerStatus returns formatted player status
func (gs *GameState) GetPlayerStatus() string {
	var result strings.Builder

	result.WriteString("\n=== 状态 ===\n")
	result.WriteString(fmt.Sprintf("生命: %d/%d", gs.Player.Health, gs.Player.MaxHealth))

	if gs.Player.HasEquippedWeapon() {
		if weapon, exists := gs.Items[gs.Player.GetEquippedWeaponID()]; exists {
			result.WriteString(fmt.Sprintf(" | 武器: %s (+%d)",
				gs.GetDisplayName(weapon.Name),
				weapon.GetDamageBonus()))
		}
	}

	result.WriteString("\n")
	return result.String()
}

// GetInventoryDisplay returns formatted inventory
func (gs *GameState) GetInventoryDisplay() string {
	if len(gs.Player.Inventory) == 0 {
		return "背包是空的。"
	}

	var result strings.Builder
	result.WriteString("\n=== 背包 ===\n")

	for _, itemID := range gs.Player.Inventory {
		if item, exists := gs.Items[itemID]; exists {
			equipped := ""
			if gs.Player.GetEquippedWeaponID() == itemID {
				equipped = " [武器]"
			}
			if gs.Player.GetEquippedArmorID() == itemID {
				equipped = " [护甲]"
			}
			result.WriteString(fmt.Sprintf("  - %s%s\n",
				gs.GetDisplayName(item.Name), equipped))
		}
	}

	result.WriteString(fmt.Sprintf("\n共 %d 件物品\n", len(gs.Player.Inventory)))
	return result.String()
}
