package models

// EnvironmentalEffects represents gameplay modifiers for a biome
// All fields are pointers to allow for nil values (no effect)
type EnvironmentalEffects struct {
	AccuracyModifier *float64 `json:"accuracy_modifier"` // Multiplier (0.8 = -20% accuracy)
	DodgeModifier    *float64 `json:"dodge_modifier"`    // Multiplier for dodge chance
	DamageModifier   *float64 `json:"damage_modifier"`   // Multiplier for damage dealt/received
	SpecialCondition *string  `json:"special_condition"` // e.g., "darkness", "slippery", "poisonous"
}

// Biome represents an environmental region with gameplay effects
type Biome struct {
	ID                   string                `json:"id"`
	Name                 Text                  `json:"name"`
	AmbientDescription   Text                  `json:"ambient_description"`
	EnvironmentalEffects *EnvironmentalEffects `json:"environmental_effects"` // nil means no effects
}

// HasEnvironmentalEffect checks if a biome has any environmental effects
func (b *Biome) HasEnvironmentalEffect() bool {
	return b.EnvironmentalEffects != nil
}

// ApplyEnvironmentalEffects modifies combat parameters based on biome
// This returns modified values (to be used in combat system)
func (b *Biome) ApplyEnvironmentalEffects(baseAccuracy, baseDodge, baseDamage float64) (accuracy, dodge, damage float64) {
	accuracy = baseAccuracy
	dodge = baseDodge
	damage = baseDamage

	if b.EnvironmentalEffects == nil {
		return
	}

	if b.EnvironmentalEffects.AccuracyModifier != nil {
		accuracy = baseAccuracy * *b.EnvironmentalEffects.AccuracyModifier
	}

	if b.EnvironmentalEffects.DodgeModifier != nil {
		dodge = baseDodge * *b.EnvironmentalEffects.DodgeModifier
	}

	if b.EnvironmentalEffects.DamageModifier != nil {
		damage = baseDamage * *b.EnvironmentalEffects.DamageModifier
	}

	return
}

// GetSpecialCondition returns the special condition string if it exists
func (b *Biome) GetSpecialCondition() string {
	if b.EnvironmentalEffects != nil && b.EnvironmentalEffects.SpecialCondition != nil {
		return *b.EnvironmentalEffects.SpecialCondition
	}
	return ""
}

// GetAccuracyModifier returns the accuracy modifier if it exists
func (b *Biome) GetAccuracyModifier() *float64 {
	if b.EnvironmentalEffects != nil {
		return b.EnvironmentalEffects.AccuracyModifier
	}
	return nil
}

// GetDodgeModifier returns the dodge modifier if it exists
func (b *Biome) GetDodgeModifier() *float64 {
	if b.EnvironmentalEffects != nil {
		return b.EnvironmentalEffects.DodgeModifier
	}
	return nil
}

// GetDamageModifier returns the damage modifier if it exists
func (b *Biome) GetDamageModifier() *float64 {
	if b.EnvironmentalEffects != nil {
		return b.EnvironmentalEffects.DamageModifier
	}
	return nil
}
