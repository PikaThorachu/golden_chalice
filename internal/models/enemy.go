type GameState struct {
    // Player data
    PlayerHealth int
    PlayerLocation string
    PlayerInventory []string  // item IDs
    
    // World state
    DefeatedEnemies map[string]bool  // location ID -> defeated
    TakenItems map[string]bool  // location ID -> taken
    PendingDrops map[string][]string  // location ID -> item IDs available to take
    
    // Display preferences
    DisplayPrefs DisplayPreferences
}