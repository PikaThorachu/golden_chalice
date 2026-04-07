package models

import (
	"math/rand"
)

// AttackRange represents min and max damage for enemy attacks
type AttackRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// EnemyDrop represents an item that can drop from an enemy
type EnemyDrop struct {
	ItemID string `json:"item_id"`
	Chance int    `json:"chance"` // 0-100 percentage
}

// SpecialAbility represents a special combat ability (future expansion)
// All fields are pointers to allow for nil values
type SpecialAbility struct {
	Name             *string `json:"name"`               // Name of the ability
	Description      *string `json:"description"`        // Description of the ability
	DamageBonus      *int    `json:"damage_bonus"`       // Additional damage
	EffectType       *string `json:"effect_type"`        // e.g., "poison", "stun", "bleed"
	EffectDuration   *int    `json:"effect_duration"`    // Turns effect lasts
	ChanceToActivate *int    `json:"chance_to_activate"` // 0-100 percentage
}

// Enemy represents a creature the player can fight
type Enemy struct {
	ID               string          `json:"id"`
	Name             Text            `json:"name"`
	Health           int             `json:"health"`
	AttackPower      AttackRange     `json:"attack_power"`
	Defense          int             `json:"defense"`
	ExperiencePoints int             `json:"experience_points"`
	Drops            []EnemyDrop     `json:"drops"`
	BiomeAffinity    []string        `json:"biome_affinity"`
	SpecialAbilities *SpecialAbility `json:"special_abilities"` // nil means no special abilities
}

// CalculateDamage calculates the damage an enemy deals in combat
func (e *Enemy) CalculateDamage() int {
	if e.AttackPower.Min == e.AttackPower.Max {
		return e.AttackPower.Min
	}

	// Random damage between min and max (inclusive)
	damage := rand.Intn(e.AttackPower.Max-e.AttackPower.Min+1) + e.AttackPower.Min

	return damage
}

// IsAlive checks if the enemy is alive based on health
func (e *Enemy) IsAlive() bool {
	return e.Health > 0
}

// TakeDamage reduces enemy health and returns the actual damage taken
func (e *Enemy) TakeDamage(amount int) int {
	if amount <= 0 {
		return 0
	}

	actualDamage := amount
	if e.Health-amount < 0 {
		actualDamage = e.Health
		e.Health = 0
	} else {
		e.Health -= amount
	}

	return actualDamage
}

// GetDropItems determines which items drop based on drop chances
// Returns slice of item IDs that dropped
func (e *Enemy) GetDropItems() []string {
	var drops []string

	for _, drop := range e.Drops {
		// Random chance (0-99) vs drop chance
		if rand.Intn(100) < drop.Chance {
			drops = append(drops, drop.ItemID)
		}
	}

	return drops
}

// HasSpecialAbility checks if the enemy has special abilities
func (e *Enemy) HasSpecialAbility() bool {
	return e.SpecialAbilities != nil
}

// ShouldActivateSpecialAbility checks if special ability should trigger
// Returns true based on chance_to_activate
func (e *Enemy) ShouldActivateSpecialAbility() bool {
	if e.SpecialAbilities == nil || e.SpecialAbilities.ChanceToActivate == nil {
		return false
	}

	chance := *e.SpecialAbilities.ChanceToActivate
	return rand.Intn(100) < chance
}

// GetSpecialAbilityName returns the name of the special ability
func (e *Enemy) GetSpecialAbilityName() string {
	if e.SpecialAbilities == nil || e.SpecialAbilities.Name == nil {
		return ""
	}
	return *e.SpecialAbilities.Name
}

// GetSpecialAbilityDescription returns the description of the special ability
func (e *Enemy) GetSpecialAbilityDescription() string {
	if e.SpecialAbilities == nil || e.SpecialAbilities.Description == nil {
		return ""
	}
	return *e.SpecialAbilities.Description
}

// GetSpecialAbilityDamageBonus returns the damage bonus of the special ability
func (e *Enemy) GetSpecialAbilityDamageBonus() int {
	if e.SpecialAbilities == nil || e.SpecialAbilities.DamageBonus == nil {
		return 0
	}
	return *e.SpecialAbilities.DamageBonus
}

// GetSpecialAbilityEffectType returns the effect type of the special ability
func (e *Enemy) GetSpecialAbilityEffectType() string {
	if e.SpecialAbilities == nil || e.SpecialAbilities.EffectType == nil {
		return ""
	}
	return *e.SpecialAbilities.EffectType
}

// GetSpecialAbilityEffectDuration returns the effect duration of the special ability
func (e *Enemy) GetSpecialAbilityEffectDuration() int {
	if e.SpecialAbilities == nil || e.SpecialAbilities.EffectDuration == nil {
		return 0
	}
	return *e.SpecialAbilities.EffectDuration
}

// AppearsInBiome checks if the enemy appears in the given biome
func (e *Enemy) AppearsInBiome(biomeID string) bool {
	for _, biome := range e.BiomeAffinity {
		if biome == biomeID {
			return true
		}
	}
	return false
}

// GetDefenseValue returns the enemy's defense (for damage calculation)
func (e *Enemy) GetDefenseValue() int {
	return e.Defense
}

// GetExperienceReward returns experience points for defeating the enemy
func (e *Enemy) GetExperienceReward() int {
	return e.ExperiencePoints
}

// GetCurrentHealth returns the current health (for display)
func (e *Enemy) GetCurrentHealth() int {
	return e.Health
}

// GetMaxHealth returns the maximum health (needs to be stored or passed)
// Note: This requires storing max health separately if needed
// For now, we'll assume current health is the max at creation
// You may want to add a MaxHealth field to the Enemy struct
