package game

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"golden_chalice/internal/errors"
	"golden_chalice/internal/logging"
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
	DefeatedEnemies map[string]bool // Make sure this line exists
	TakenItems      map[string]bool
	PendingDrops    map[string][]string
	GameOver        bool
	GameWon         bool

	// Room tracking
	CurrentRoomID  *string `json:"current_room_id"`  // Track current room (nil if not in a room)
	LastLocationID string  `json:"last_location_id"` // Track last location for room entry detection	DefeatedEnemies map[string]bool

	// Formatter and Logger
	Formatter *DisplayFormatter
	Logger    *logging.Logger
}

// NewGameState creates a new game state with loaded static data
func NewGameState(
	config *models.Config,
	world *models.World,
	items map[string]models.Item,
	enemies map[string]models.Enemy,
	biomes map[string]models.Biome,
	logger *logging.Logger,
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
		Formatter:       NewDisplayFormatter(config),
		Logger:          logger,
		CurrentRoomID:   nil,
		LastLocationID:  "",
	}
}

// SafeMove attempts to move with proper error handling
func (gs *GameState) SafeMove(direction models.Direction) error {
	// Check if game is active
	if gs.GameOver {
		gs.Logger.LogError(errors.ErrGameAlreadyOver, map[string]interface{}{"direction": direction.String()})
		return errors.ErrGameAlreadyOver
	}

	// Check if player is alive
	if !gs.Player.IsAlive() {
		gs.Logger.LogError(errors.ErrPlayerDead, map[string]interface{}{"player_health": gs.Player.Health})
		return errors.ErrPlayerDead
	}

	// Check for enemy blocking
	if gs.HasEnemyAtCurrentLocation() {
		enemy, _ := gs.GetCurrentEnemy()
		gs.Logger.LogMovement(gs.Player.CurrentLocationID, "", direction.String())
		return errors.Wrap(errors.ErrEnemyBlocks, errors.ErrTypeMovement,
			fmt.Sprintf("有%s挡住了去路", gs.GetDisplayName(enemy.Name)),
			fmt.Sprintf("You %s dang zhu le qu lu", enemy.Name.Pinyin),
			fmt.Sprintf("%s blocks your path", enemy.Name.English))
	}

	// Log movement attempt
	oldLocationID := gs.Player.CurrentLocationID
	gs.Logger.LogMovement(oldLocationID, "unknown", direction.String())

	// Attempt movement
	destID, err := gs.World.GetDestination(
		gs.Player.CurrentLocationID,
		direction,
		gs.Player.Inventory,
	)
	if err != nil {
		if strings.Contains(err.Error(), "需要") {
			gs.Logger.LogError(errors.ErrExitLocked, map[string]interface{}{"direction": direction.String()})
			return errors.Wrap(errors.ErrExitLocked, errors.ErrTypeMovement, err.Error(), err.Error(), err.Error())
		}
		gs.Logger.LogError(errors.ErrNoExit, map[string]interface{}{"direction": direction.String()})
		return errors.Wrap(errors.ErrNoExit, errors.ErrTypeMovement, err.Error(), err.Error(), err.Error())
	}

	// Move player
	gs.Player.CurrentLocationID = destID
	gs.LastLocationID = oldLocationID

	// Check if we entered a new room
	gs.updateCurrentRoom()

	gs.Logger.LogMovement(oldLocationID, destID, direction.String())
	return nil
}

// SafeTakeItem attempts to take an item with proper error handling
func (gs *GameState) SafeTakeItem(itemID string) error {
	if gs.GameOver {
		gs.Logger.LogError(errors.ErrGameAlreadyOver, map[string]interface{}{"item": itemID})
		return errors.ErrGameAlreadyOver
	}

	if !gs.Player.IsAlive() {
		gs.Logger.LogError(errors.ErrPlayerDead, map[string]interface{}{"item": itemID})
		return errors.ErrPlayerDead
	}

	if gs.HasEnemyAtCurrentLocation() {
		gs.Logger.LogError(errors.ErrEnemyBlocks, map[string]interface{}{"item": itemID})
		return errors.ErrEnemyBlocks
	}

	// Check inventory capacity before taking
	if gs.Player.IsInventoryFull() {
		gs.Logger.LogError(errors.ErrInventoryFull, map[string]interface{}{"item": itemID, "capacity": len(gs.Player.Inventory)})
		return errors.ErrInventoryFull
	}

	err := gs.TakeItem(itemID)
	if err != nil {
		gs.Logger.LogError(errors.ErrItemNotFound, map[string]interface{}{"item": itemID})
		return errors.Wrap(errors.ErrItemNotFound, errors.ErrTypeInventory, err.Error(), err.Error(), err.Error())
	}

	gs.Logger.LogPlayerAction("take", map[string]interface{}{"item": itemID, "location": gs.Player.CurrentLocationID})
	return nil
}

// SafeUseItem attempts to use an item with proper error handling
func (gs *GameState) SafeUseItem(itemID string) (string, error) {
	if gs.GameOver {
		gs.Logger.LogError(errors.ErrGameAlreadyOver, map[string]interface{}{"item": itemID})
		return "", errors.ErrGameAlreadyOver
	}

	if !gs.Player.IsAlive() {
		gs.Logger.LogError(errors.ErrPlayerDead, map[string]interface{}{"item": itemID})
		return "", errors.ErrPlayerDead
	}

	result, err := gs.UseItem(itemID)
	if err != nil {
		gs.Logger.LogError(errors.ErrItemNotFound, map[string]interface{}{"item": itemID})
		return "", errors.Wrap(errors.ErrItemNotFound, errors.ErrTypeInventory, err.Error(), err.Error(), err.Error())
	}

	gs.Logger.LogPlayerAction("use", map[string]interface{}{"item": itemID})
	return result, nil
}

// SafeEquipBackpack attempts to equip a backpack
func (gs *GameState) SafeEquipBackpack(itemID string) (string, error) {
	if gs.GameOver {
		return "", errors.ErrGameAlreadyOver
	}

	if !gs.Player.IsAlive() {
		return "", errors.ErrPlayerDead
	}

	if !gs.Player.HasItem(itemID) {
		return "", errors.ErrItemNotFound
	}

	item, exists := gs.Items[itemID]
	if !exists {
		return "", errors.ErrItemNotFound
	}

	if !item.IsBackpack() {
		return "", errors.New(errors.ErrTypeInventory,
			"这不是一个背包",
			"Zhe bu shi yi ge bei bao",
			"This is not a backpack")
	}

	sizeBonus := item.GetSizeBonus()
	maxSize := gs.Config.GetMaxInventorySize()

	success, increase, message := gs.Player.EquipBackpack(itemID, sizeBonus, maxSize)
	if !success {
		return "", errors.New(errors.ErrTypeInventory, message, message, message)
	}

	if gs.Logger != nil {
		gs.Logger.LogPlayerAction("equip_backpack", map[string]interface{}{
			"item":         itemID,
			"size_bonus":   sizeBonus,
			"new_capacity": gs.Player.InventorySize,
		})
	}

	// Natural Chinese for backpack equipping
	return gs.Formatter.formatInline(
		fmt.Sprintf("带上了 %s，背包容量 +%d (当前: %d)", item.Name.Chinese, increase, gs.Player.InventorySize),
		fmt.Sprintf("Dai shang le %s，bei bao rong liang +%d (dang qian: %d)", item.Name.Pinyin, increase, gs.Player.InventorySize),
		fmt.Sprintf("Equipped %s, inventory +%d (current: %d)", item.Name.English, increase, gs.Player.InventorySize),
	), nil
}

// SafeUnequipBackpack attempts to unequip the current backpack
func (gs *GameState) SafeUnequipBackpack() (string, error) {
	if gs.GameOver {
		return "", errors.ErrGameAlreadyOver
	}

	if !gs.Player.HasBackpackEquipped() {
		return "", errors.New(errors.ErrTypeInventory,
			"没有装备背包",
			"Méiyǒu zhuāngbèi bēibāo",
			"No backpack equipped")
	}

	// Get the backpack item to know the size bonus
	backpackID := gs.Player.GetEquippedBackpackID()
	backpack, exists := gs.Items[backpackID]
	if !exists {
		// Force unequip without bonus calculation
		gs.Player.UnequipBackpack()
		return "卸下了背包（数据异常）", nil
	}

	// Calculate what the size should be without the backpack
	sizeBonus := backpack.GetSizeBonus()
	newSize := gs.Player.InventorySize - sizeBonus

	// Ensure new size is at least current item count
	if newSize < len(gs.Player.Inventory) {
		return "", errors.New(errors.ErrTypeInventory,
			fmt.Sprintf("背包中有太多物品，无法卸下背包 (当前: %d/%d)", len(gs.Player.Inventory), gs.Player.InventorySize),
			fmt.Sprintf("Bēibāo zhōng yǒu tài duō wùpǐn, wúfǎ xièxià bēibāo (dāngqián: %d/%d)", len(gs.Player.Inventory), gs.Player.InventorySize),
			fmt.Sprintf("Too many items in inventory to unequip backpack (current: %d/%d)", len(gs.Player.Inventory), gs.Player.InventorySize))
	}

	success, decrease, message := gs.Player.UnequipBackpack()
	if !success {
		return "", errors.New(errors.ErrTypeInventory, message, message, message)
	}

	gs.Logger.LogPlayerAction("unequip_backpack", map[string]interface{}{
		"item":          backpackID,
		"size_decrease": decrease,
		"new_capacity":  gs.Player.InventorySize,
	})

	return gs.Formatter.formatInline(
		fmt.Sprintf("卸下了 %s，背包容量 -%d (当前: %d)", backpack.Name.Chinese, decrease, gs.Player.InventorySize),
		fmt.Sprintf("Xièxià le %s，bēibāo róngliàng -%d (dāngqián: %d)", backpack.Name.Pinyin, decrease, gs.Player.InventorySize),
		fmt.Sprintf("Unequipped %s, inventory -%d (current: %d)", backpack.Name.English, decrease, gs.Player.InventorySize),
	), nil
}

// ProcessCombatTurn handles one turn of combat with logging
func (gs *GameState) ProcessCombatTurn(enemy *models.Enemy, playerAttack bool) (bool, bool, string) {
	if playerAttack {
		// Player attacks
		baseDamage := gs.Config.CombatSettings.BaseDamageMin
		damage := gs.Player.GetAttackDamage(baseDamage, gs.Items)
		actualDamage := enemy.TakeDamage(damage)

		gs.Logger.LogCombat(gs.Player.Name, gs.GetDisplayName(enemy.Name), actualDamage, "player_attack")

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
			gs.Logger.LogCombat(gs.Player.Name, gs.GetDisplayName(enemy.Name), actualDamage, "victory")
			return true, false, fmt.Sprintf("你对 %s 造成了 %d 点伤害！击败了敌人！获得 %d 点经验值。",
				gs.GetDisplayName(enemy.Name), actualDamage, exp)
		}

		return false, false, fmt.Sprintf("你对 %s 造成了 %d 点伤害！", gs.GetDisplayName(enemy.Name), actualDamage)
	}

	// Enemy attacks
	damage := enemy.CalculateDamage()
	actualDamage := gs.Player.TakeDamage(damage)

	gs.Logger.LogCombat(gs.GetDisplayName(enemy.Name), gs.Player.Name, actualDamage, "enemy_attack")

	if !gs.Player.IsAlive() {
		gs.Logger.LogCombat(gs.GetDisplayName(enemy.Name), gs.Player.Name, actualDamage, "player_defeated")
		return false, true, fmt.Sprintf("%s 对你造成了 %d 点伤害！你被击败了...",
			gs.GetDisplayName(enemy.Name), actualDamage)
	}

	return false, false, fmt.Sprintf("%s 对你造成了 %d 点伤害！", gs.GetDisplayName(enemy.Name), actualDamage)
}

// AddPendingDrop adds an item to the ground at a location with logging
func (gs *GameState) AddPendingDrop(locationID string, itemID string) {
	gs.PendingDrops[locationID] = append(gs.PendingDrops[locationID], itemID)
	gs.Logger.LogPlayerAction("drop", map[string]interface{}{
		"item":     itemID,
		"location": locationID,
	})
}

// GetDisplayName returns formatted name based on config preferences
func (gs *GameState) GetDisplayName(text models.Text) string {
	return gs.Formatter.FormatText(text)
}

// GetDisplayDescription returns formatted description based on config preferences
func (gs *GameState) GetDisplayDescription(text models.Text) string {
	return gs.Formatter.FormatText(text)
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

	var items []string
	if !gs.TakenItems[room.ID] {
		items = append(items, room.ItemIDs...)
	}

	pendingItems := gs.PendingDrops[room.ID]
	items = append(items, pendingItems...)

	return items
}

// TakeItem removes an item from the current location and adds to player inventory
func (gs *GameState) TakeItem(itemID string) error {
	room, err := gs.GetCurrentRoom()
	if err != nil {
		return err
	}

	if slices.Contains(room.ItemIDs, itemID) {
		if gs.TakenItems[room.ID] {
			return fmt.Errorf("这个物品已经被拿走了")
		}

		success, errMsg := gs.Player.AddItemWithCheck(itemID)
		if !success {
			return fmt.Errorf(errMsg)
		}
		gs.TakenItems[room.ID] = true
		return nil
	}

	pendingItems := gs.PendingDrops[room.ID]
	for i, id := range pendingItems {
		if id == itemID {
			success, errMsg := gs.Player.AddItemWithCheck(itemID)
			if !success {
				return fmt.Errorf(errMsg)
			}
			gs.PendingDrops[room.ID] = append(pendingItems[:i], pendingItems[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("这里没有 %s", itemID)
}

// Move attempts to move the player in a direction (legacy method, prefer SafeMove)
func (gs *GameState) Move(direction models.Direction) error {
	return gs.SafeMove(direction)
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

	// Check if player has golden chalice
	if !gs.Player.HasItem("golden_chalice") {
		return false
	}

	// Check if player is in Cave Entrance room
	currentLocation, err := gs.GetCurrentRoom()
	if err != nil {
		return false
	}

	currentRoom, err := gs.World.GetCurrentRoom(currentLocation.ID)
	if err != nil || currentRoom == nil {
		return false
	}

	if currentRoom.ID == "entrance_complex" {
		gs.GameWon = true
		gs.GameOver = true
		if gs.Logger != nil {
			gs.Logger.LogGameEvent("victory", map[string]interface{}{
				"player":   gs.Player.Name,
				"location": currentLocation.ID,
			})
		}
		return true
	}

	return false
}

// CheckLossCondition checks if player has lost the game
func (gs *GameState) CheckLossCondition() bool {
	if !gs.Player.IsAlive() {
		gs.GameOver = true
		gs.Logger.LogGameEvent("game_over", map[string]interface{}{
			"player": gs.Player.Name,
			"health": gs.Player.Health,
		})
		return true
	}
	return false
}

// GetCurrentRoomDescription returns the room description if entering a new room
func (gs *GameState) GetCurrentRoomDescription() string {
	// Check if we just entered a new room
	currentRoom, err := gs.World.GetCurrentRoom(gs.Player.CurrentLocationID)
	if err != nil {
		return gs.getLocationDescription()
	}

	var currentRoomID *string
	if currentRoom != nil {
		currentRoomID = &currentRoom.ID
	}

	// If entering a new room, show room description
	if (gs.CurrentRoomID == nil && currentRoomID != nil) ||
		(gs.CurrentRoomID != nil && currentRoomID != nil && *gs.CurrentRoomID != *currentRoomID) ||
		gs.LastLocationID == "" { // First entry
		gs.CurrentRoomID = currentRoomID
		if currentRoom != nil {
			return gs.getRoomEntryDescription(currentRoom)
		}
	}

	// Otherwise just show location description
	return gs.getLocationDescription()
}

// getRoomEntryDescription returns the room description when first entering
func (gs *GameState) getRoomEntryDescription(room *models.Room) string {
	var result strings.Builder

	// Room name with decoration
	result.WriteString(gs.Formatter.FormatRoomTitle(room.Name))

	// Room description
	result.WriteString(gs.Formatter.FormatDescription(room.Description))
	result.WriteString("\n\n")

	// Then show current location details
	result.WriteString(gs.getLocationDescription())

	return result.String()
}

// GetPlayerStatus returns formatted player status
func (gs *GameState) GetPlayerStatus() string {
	var result strings.Builder

	result.WriteString(gs.Formatter.FormatHeader(models.Text{
		Chinese: "角色状态",
		Pinyin:  "Juésè Zhuàngtài",
		English: "Character Status",
	}))

	result.WriteString(fmt.Sprintf("%s: %s\n",
		gs.Formatter.formatInline("名称", "Míngchēng", "Name"),
		gs.Player.Name))

	result.WriteString(fmt.Sprintf("%s: %d\n",
		gs.Formatter.formatInline("等级", "Děngjí", "Level"),
		gs.Player.Level))

	healthText := gs.Formatter.FormatHealthStatus(gs.Player.Health, gs.Player.MaxHealth)
	healthBar := gs.Formatter.FormatHealthBar(gs.Player.Health, gs.Player.MaxHealth, 20)
	result.WriteString(fmt.Sprintf("%s: %d/%d %s %s\n",
		gs.Formatter.formatInline("生命", "Shēngmìng", "Health"),
		gs.Player.Health, gs.Player.MaxHealth,
		healthText, healthBar))

	expNeeded := gs.Player.Level * 100
	result.WriteString(fmt.Sprintf("%s: %d / %d\n",
		gs.Formatter.formatInline("经验", "Jīngyàn", "Experience"),
		gs.Player.ExperiencePoints, expNeeded))

	// Inventory capacity
	currentInv, maxInv := gs.Player.GetInventoryCapacity()
	result.WriteString(fmt.Sprintf("%s: %d / %d\n",
		gs.Formatter.formatInline("背包容量", "Bēibāo róngliàng", "Inventory"),
		currentInv, maxInv))

	result.WriteString(gs.Formatter.FormatSubHeader(models.Text{
		Chinese: "装备",
		Pinyin:  "Zhuāngbèi",
		English: "Equipment",
	}))

	if gs.Player.HasEquippedWeapon() {
		if weapon, exists := gs.Items[gs.Player.GetEquippedWeaponID()]; exists {
			result.WriteString(fmt.Sprintf("  %s: %s (+%d)\n",
				gs.Formatter.formatInline("武器", "Wǔqì", "Weapon"),
				gs.GetDisplayName(weapon.Name),
				weapon.GetDamageBonus()))
		}
	} else {
		result.WriteString(fmt.Sprintf("  %s: %s\n",
			gs.Formatter.formatInline("武器", "Wǔqì", "Weapon"),
			gs.Formatter.formatInline("无", "Wú", "None")))
	}

	if gs.Player.HasEquippedArmor() {
		if armor, exists := gs.Items[gs.Player.GetEquippedArmorID()]; exists {
			result.WriteString(fmt.Sprintf("  %s: %s (+%d)\n",
				gs.Formatter.formatInline("护甲", "Hùjiǎ", "Armor"),
				gs.GetDisplayName(armor.Name),
				armor.GetDefenseBonus()))
		}
	} else {
		result.WriteString(fmt.Sprintf("  %s: %s\n",
			gs.Formatter.formatInline("护甲", "Hùjiǎ", "Armor"),
			gs.Formatter.formatInline("无", "Wú", "None")))
	}

	room, err := gs.GetCurrentRoom()
	if err == nil {
		result.WriteString(gs.Formatter.FormatSubHeader(models.Text{
			Chinese: "当前位置",
			Pinyin:  "Dāngqián Wèizhì",
			English: "Current Location",
		}))
		result.WriteString(fmt.Sprintf("  %s\n", gs.GetDisplayName(room.Name)))
	}

	return result.String()
}

// GetInventoryDisplay returns formatted inventory
func (gs *GameState) GetInventoryDisplay() string {
	if len(gs.Player.Inventory) == 0 {
		return gs.Formatter.formatInline(
			"你的背包是空的。",
			"Nǐ de bēibāo shì kōng de.",
			"Your backpack is empty.",
		)
	}

	var result strings.Builder
	result.WriteString(gs.Formatter.FormatHeader(models.Text{
		Chinese: "背包",
		Pinyin:  "Bēibāo",
		English: "Inventory",
	}))

	for _, itemID := range gs.Player.Inventory {
		if item, exists := gs.Items[itemID]; exists {
			equipped := ""
			if gs.Player.GetEquippedWeaponID() == itemID {
				equipped = gs.Formatter.formatInline(" [已装备]", " [yǐ zhuāngbèi]", " [equipped]")
			}
			if gs.Player.GetEquippedArmorID() == itemID {
				equipped = gs.Formatter.formatInline(" [已装备]", " [yǐ zhuāngbèi]", " [equipped]")
			}
			result.WriteString(fmt.Sprintf("  • %s%s\n", gs.GetDisplayName(item.Name), equipped))
		}
	}

	currentInv, maxInv := gs.Player.GetInventoryCapacity()
	result.WriteString(gs.Formatter.formatInline(
		fmt.Sprintf("\n容量: %d / %d", currentInv, maxInv),
		fmt.Sprintf("Róngliàng: %d / %d", currentInv, maxInv),
		fmt.Sprintf("Capacity: %d / %d", currentInv, maxInv),
	))

	return result.String()
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

// NewGame creates a new player and starts a fresh game
func (gs *GameState) NewGame(playerName string) {
	gs.Player = models.NewPlayer(
		playerName,
		gs.Config.GetStartingLocationID(),
		gs.Config.GetStartingHealth(),
		gs.Config.GetStartingInventorySize(),
	)
	gs.DefeatedEnemies = make(map[string]bool)
	gs.TakenItems = make(map[string]bool)
	gs.PendingDrops = make(map[string][]string)
	gs.GameOver = false
	gs.GameWon = false
	gs.CurrentRoomID = nil
	gs.LastLocationID = ""

	if gs.Logger != nil {
		gs.Logger.LogGameEvent("new_game", map[string]interface{}{
			"player":   playerName,
			"location": gs.Config.GetStartingLocationID(),
		})
	}
}

// AttemptFlee attempts to flee from combat
// Returns (success, message)
func (gs *GameState) AttemptFlee() (bool, string) {
	fleeRate := gs.Config.CombatSettings.FleeSuccessRate

	// Use random chance for flee
	chance := rand.Intn(100)
	if chance < fleeRate {
		return true, gs.Formatter.formatInline(
			"你成功逃跑了！",
			"Ni cheng gong tao pao le!",
			"You successfully fled!",
		)
	}

	return false, gs.Formatter.formatInline(
		"逃跑失败！",
		"Tao pao shi bai!",
		"Failed to flee!",
	)
}

// GetLookDescription returns the full look description including room description
func (gs *GameState) GetLookDescription() string {
	var result strings.Builder

	// Show room description if in a room
	currentRoom, err := gs.World.GetCurrentRoom(gs.Player.CurrentLocationID)
	if err == nil && currentRoom != nil {
		result.WriteString(gs.Formatter.FormatRoomTitle(currentRoom.Name))
		result.WriteString(gs.Formatter.FormatDescription(currentRoom.Description))
		result.WriteString("\n\n")
	}

	// Show current location info
	result.WriteString(gs.getLocationDescription())

	return result.String()
}

// getLocationDescription returns just the location description (no room header)
func (gs *GameState) getLocationDescription() string {
	location, err := gs.GetCurrentRoom()
	if err != nil {
		return "无法获取位置描述"
	}

	var result strings.Builder

	// Location name (only show if not in a room or as subheader)
	if gs.CurrentRoomID == nil {
		result.WriteString(gs.Formatter.FormatRoomTitle(location.Name))
	} else {
		result.WriteString(gs.Formatter.FormatSubHeader(location.Name))
	}

	// Note: Location descriptions have been removed - rooms provide the atmospheric description
	// Only show biome ambient description and items/exits

	// Biome ambient description
	biome, err := gs.GetCurrentBiome()
	if err == nil && (biome.AmbientDescription.Chinese != "" || biome.AmbientDescription.English != "") {
		result.WriteString(gs.Formatter.FormatDescription(biome.AmbientDescription))
		result.WriteString("\n\n")
	}

	// Items in location
	items := gs.GetItemsAtCurrentLocation()
	if len(items) > 0 {
		itemNames := make([]string, len(items))
		for i, itemID := range items {
			if item, exists := gs.Items[itemID]; exists {
				itemNames[i] = gs.GetDisplayName(item.Name)
			}
		}
		result.WriteString(gs.Formatter.FormatItemList(itemNames))
		result.WriteString("\n\n")
	}

	// Available exits
	exits := gs.GetAvailableExits()
	result.WriteString(gs.Formatter.FormatExitList(exits))
	result.WriteString("\n")

	return result.String()
}

// updateCurrentRoom checks if the player has entered a new room and updates state
func (gs *GameState) updateCurrentRoom() {
	newRoom, err := gs.World.GetCurrentRoom(gs.Player.CurrentLocationID)
	if err != nil {
		return
	}

	oldRoomID := gs.CurrentRoomID
	var newRoomID *string
	if newRoom != nil {
		newRoomID = &newRoom.ID
	}

	gs.CurrentRoomID = newRoomID

	// Log room entry if changed
	if gs.Logger != nil && newRoom != nil &&
		((oldRoomID == nil && newRoomID != nil) ||
			(oldRoomID != nil && newRoomID != nil && *oldRoomID != *newRoomID)) {
		gs.Logger.LogGameEvent("enter_room", map[string]interface{}{
			"room_id":       newRoom.ID,
			"room_name":     newRoom.Name.Chinese,
			"from_location": gs.LastLocationID,
			"to_location":   gs.Player.CurrentLocationID,
		})
	}
}
