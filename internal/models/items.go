package models

import (
	"fmt"
)

// ItemType represents the type of item
type ItemType string

const (
	ItemTypeWeapon     ItemType = "weapon"
	ItemTypeQuest      ItemType = "quest"
	ItemTypeConsumable ItemType = "consumable"
	ItemTypeKey        ItemType = "key"
	ItemTypeArmor      ItemType = "armor"
	ItemTypeBackpack   ItemType = "backpack" // Add this line
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
