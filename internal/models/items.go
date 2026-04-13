package models

import (
	"fmt"
)

// ItemType represents the type of item
type ItemType string

const (
	ItemTypeWeapon      ItemType = "weapon"
	ItemTypeQuest       ItemType = "quest"
	ItemTypeConsumable  ItemType = "consumable"
	ItemTypeKey         ItemType = "key"
	ItemTypeArmor       ItemType = "armor"
	ItemTypeBackpack    ItemType = "backpack"
	ItemTypeInspectable ItemType = "inspectable"
	ItemTypeJunk        ItemType = "junk"
	ItemTypeContainer   ItemType = "container"
)

// ItemProperties contains type-specific properties
// All fields are pointers to allow for nil values (property not applicable)
type ItemProperties struct {
	// Weapon properties
	DamageBonus *int  `json:"damage_bonus"` // Extra damage when equipped
	Equippable  *bool `json:"equippable"`   // Can be equipped as weapon

	// Consumable properties
	HealthRestore  *int `json:"health_restore"`  // HP restored when used
	ManaRestore    *int `json:"mana_restore"`    // MP restored (for future)
	EffectDuration *int `json:"effect_duration"` // Turns effect lasts

	// Quest properties
	WinCondition *bool   `json:"win_condition"` // Triggers game victory
	StoryFlag    *string `json:"story_flag"`    // Quest progression marker

	// Key properties
	OpensDoorID *string `json:"opens_door_id"` // Specific door this key opens

	// Armor properties
	DefenseBonus *int    `json:"defense_bonus"` // Extra defense when equipped
	EquipSlot    *string `json:"equip_slot"`    // "head", "chest", "hands", etc.

	// Backpack properties (add this section)
	SizeBonus *int `json:"size_bonus"` // Additional inventory slots

	// Inspectable properties (add these)
	ContainsItemID *string `json:"contains_item_id"` // Item ID found inside
	TrapEnemyID    *string `json:"trap_enemy_id"`    // Enemy ID that spawns on inspect
	InspectMessage *string `json:"inspect_message"`  // Message shown when inspecting

	// Container properties
	Capacity *int `json:"capacity"` // Maximum items this container can hold
}

// Item represents any item in the game
type Item struct {
	ID          string         `json:"id"`
	Name        Text           `json:"name"`
	Description Text           `json:"description"`
	Type        ItemType       `json:"type"`
	Properties  ItemProperties `json:"properties"`
	Usable      bool           `json:"usable"`
	Consumable  bool           `json:"consumable"`
	Inventory   []string       `json:"inventory"` // Items contained within this item
}

// IsContainer checks if the item can hold other items
func (i *Item) IsContainer() bool {
	return i.Properties.Capacity != nil && *i.Properties.Capacity > 0
}

// GetCapacity returns the maximum number of items this container can hold
func (i *Item) GetCapacity() int {
	if i.Properties.Capacity != nil {
		return *i.Properties.Capacity
	}
	return 0
}

// IsInventoryFull checks if the container's inventory is full
func (i *Item) IsInventoryFull() bool {
	return len(i.Inventory) >= i.GetCapacity()
}

// GetInventoryCount returns the number of items in the container
func (i *Item) GetInventoryCount() int {
	return len(i.Inventory)
}

// GetRemainingCapacity returns available space in the container
func (i *Item) GetRemainingCapacity() int {
	return i.GetCapacity() - len(i.Inventory)
}

// AddItemToContainer adds an item to the container's inventory
func (i *Item) AddItemToContainer(itemID string) (bool, string) {
	if !i.IsContainer() {
		return false, "这个物品不能存放其他物品"
	}
	if i.IsInventoryFull() {
		return false, fmt.Sprintf("容器已满 (容量: %d/%d)", len(i.Inventory), i.GetCapacity())
	}
	i.Inventory = append(i.Inventory, itemID)
	return true, ""
}

// RemoveItemFromContainer removes an item from the container's inventory
func (i *Item) RemoveItemFromContainer(itemID string) bool {
	for idx, id := range i.Inventory {
		if id == itemID {
			i.Inventory = append(i.Inventory[:idx], i.Inventory[idx+1:]...)
			return true
		}
	}
	return false
}

// HasItemInContainer checks if the container has a specific item
func (i *Item) HasItemInContainer(itemID string) bool {
	for _, id := range i.Inventory {
		if id == itemID {
			return true
		}
	}
	return false
}

// IsWeapon checks if the item is a weapon
func (i *Item) IsWeapon() bool {
	return i.Type == ItemTypeWeapon
}

// IsQuestItem checks if the item is a quest item
func (i *Item) IsQuestItem() bool {
	return i.Type == ItemTypeQuest
}

// IsConsumable checks if the item is consumable
func (i *Item) IsConsumable() bool {
	return i.Type == ItemTypeConsumable
}

// IsKey checks if the item is a key
func (i *Item) IsKey() bool {
	return i.Type == ItemTypeKey
}

// IsArmor checks if the item is armor
func (i *Item) IsArmor() bool {
	return i.Type == ItemTypeArmor
}

// IsUsable checks if the item can be used from inventory
func (i *Item) IsUsable() bool {
	return i.Usable
}

// IsBackpack checks if the item is a backpack
func (i *Item) IsBackpack() bool {
	return i.Type == ItemTypeBackpack
}

// IsInspectable checks if the item can be inspected
func (i *Item) IsInspectable() bool {
	return i.Type == ItemTypeInspectable
}

// IsJunk checks if the item is junk (no gameplay effect)
func (i *Item) IsJunk() bool {
	return i.Type == ItemTypeJunk
}

// GetSizeBonus returns the inventory size bonus for backpacks
func (i *Item) GetSizeBonus() int {
	if i.Type == ItemTypeBackpack && i.Properties.SizeBonus != nil {
		return *i.Properties.SizeBonus
	}
	return 0
}

// GetDamageBonus returns the damage bonus for weapons
func (i *Item) GetDamageBonus() int {
	if i.Type == ItemTypeWeapon && i.Properties.DamageBonus != nil {
		return *i.Properties.DamageBonus
	}
	return 0
}

// GetDefenseBonus returns the defense bonus for armor
func (i *Item) GetDefenseBonus() int {
	if i.Type == ItemTypeArmor && i.Properties.DefenseBonus != nil {
		return *i.Properties.DefenseBonus
	}
	return 0
}

// GetHealthRestore returns the health restored by consumables
func (i *Item) GetHealthRestore() int {
	if i.Type == ItemTypeConsumable && i.Properties.HealthRestore != nil {
		return *i.Properties.HealthRestore
	}
	return 0
}

// GetManaRestore returns the mana restored by consumables
func (i *Item) GetManaRestore() int {
	if i.Type == ItemTypeConsumable && i.Properties.ManaRestore != nil {
		return *i.Properties.ManaRestore
	}
	return 0
}

// IsWinCondition checks if the item triggers victory
func (i *Item) IsWinCondition() bool {
	if i.Type == ItemTypeQuest && i.Properties.WinCondition != nil {
		return *i.Properties.WinCondition
	}
	return false
}

// GetStoryFlag returns the story flag for quest items
func (i *Item) GetStoryFlag() string {
	if i.Type == ItemTypeQuest && i.Properties.StoryFlag != nil {
		return *i.Properties.StoryFlag
	}
	return ""
}

// GetOpensDoorID returns which door this key opens
func (i *Item) GetOpensDoorID() string {
	if i.Type == ItemTypeKey && i.Properties.OpensDoorID != nil {
		return *i.Properties.OpensDoorID
	}
	return ""
}

// GetEquipSlot returns which slot armor equips to
func (i *Item) GetEquipSlot() string {
	if i.Type == ItemTypeArmor && i.Properties.EquipSlot != nil {
		return *i.Properties.EquipSlot
	}
	return ""
}

// IsEquippable checks if weapon/armor can be equipped
func (i *Item) IsEquippable() bool {
	if i.Type == ItemTypeWeapon && i.Properties.Equippable != nil {
		return *i.Properties.Equippable
	}
	if i.Type == ItemTypeArmor {
		return true // Armor is always equippable
	}
	return false
}

// GetEffectDuration returns the effect duration for consumables
func (i *Item) GetEffectDuration() int {
	if i.Type == ItemTypeConsumable && i.Properties.EffectDuration != nil {
		return *i.Properties.EffectDuration
	}
	return 0
}

// Use applies consumable effects and returns the effect description
// This is a helper for gameplay logic
func (i *Item) Use() (healthRestored, manaRestored int, message string) {
	if !i.Usable || !i.Consumable {
		return 0, 0, fmt.Sprintf("不能使用 %s", i.Name.Chinese)
	}

	healthRestored = i.GetHealthRestore()
	manaRestored = i.GetManaRestore()

	if healthRestored > 0 && manaRestored > 0 {
		message = fmt.Sprintf("你使用了 %s，恢复了 %d 点生命值和 %d 点法力值",
			i.Name.Chinese, healthRestored, manaRestored)
	} else if healthRestored > 0 {
		message = fmt.Sprintf("你使用了 %s，恢复了 %d 点生命值",
			i.Name.Chinese, healthRestored)
	} else if manaRestored > 0 {
		message = fmt.Sprintf("你使用了 %s，恢复了 %d 点法力值",
			i.Name.Chinese, manaRestored)
	} else {
		message = fmt.Sprintf("你使用了 %s，但什么也没发生", i.Name.Chinese)
	}

	return healthRestored, manaRestored, message
}

// GetDisplayName returns the item's name in the user's preferred language
func (i *Item) GetDisplayName(config *Config) string {
	return config.GetDisplayText(i.Name)
}

// GetDisplayDescription returns the item's description in the user's preferred language
func (i *Item) GetDisplayDescription(config *Config) string {
	return config.GetDisplayText(i.Description)
}

// GetContainsItemID returns the item ID contained in this inspectable item
func (i *Item) GetContainsItemID() string {
	if i.Type == ItemTypeInspectable && i.Properties.ContainsItemID != nil {
		return *i.Properties.ContainsItemID
	}
	return ""
}

// GetTrapEnemyID returns the enemy ID that spawns when inspecting
func (i *Item) GetTrapEnemyID() string {
	if i.Type == ItemTypeInspectable && i.Properties.TrapEnemyID != nil {
		return *i.Properties.TrapEnemyID
	}
	return ""
}

// GetInspectMessage returns the message shown when inspecting
func (i *Item) GetInspectMessage() string {
	if i.Type == ItemTypeInspectable && i.Properties.InspectMessage != nil {
		return *i.Properties.InspectMessage
	}
	return ""
}
