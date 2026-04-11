Dark Caverns - Golden Chalice Adventure
A text-based RPG written in Go, featuring trilingual support (Chinese, Pinyin, English), dynamic inventory management, room-based navigation, and a robust save/load system.

Project Overview
This game is a modern reimagining of 1980s text adventures (like Zork or Colossal Cave Adventure), built with a focus on modular architecture, data-driven design, and language learning accessibility.

Core Features
Trilingual Output - All game text displays in Chinese characters, Pinyin (without tone marks for compatibility), and English

Configurable Display - Users can toggle which language(s) to display via config.json

JSON-Driven World - All game data (rooms, locations, items, enemies, biomes) loaded from JSON files

Room System - Locations grouped into Rooms; room descriptions display on entry, location names as subheaders

8-Directional Movement - Supports cardinal (N/S/E/W) and diagonal (NW/NE/SW/SE) directions

Inventory Management - Dynamic inventory capacity with backpack items that increase capacity

Equipment System - Weapons (damage bonus) and armor (defense bonus) that affect combat

Combat System - Turn-based combat with random damage, experience points, and item drops

Save/Load System - Multiple save slots, auto-save, quick-save, and save file validation

Input Validation - Comprehensive input sanitization and helpful error messages

Code Structure Decisions
Package Organization
text
internal/
├── models/      # Data structures only (what data IS)
├── loader/      # File I/O and JSON parsing (how to GET data)
├── game/        # Game logic and state management (what to DO with data)
├── save/        # Save/load functionality
├── errors/      # Structured error types with multilingual support
└── logging/     # Game logging system
Key Architectural Decisions
Separation of Models and Loaders

models/ contains only struct definitions and methods on those structs

loader/ handles file reading, JSON parsing, and validation

This prevents circular dependencies and improves testability

JSON-First Design

All game content defined in JSON files (world.json, items.json, enemies.json, etc.)

Enables modding and content updates without recompilation

Cross-referencing validation ensures data integrity

Room vs Location Distinction

Rooms - Large areas with descriptive text (displayed on entry)

Locations - Individual nodes within Rooms (displayed as subheaders)

Movement within same Room doesn't repeat room description

Direction Parsing

Supports multiple input formats: 北, 往北, 往北走, 往北去, go north, walk north

Diagonal directions: 西北, 东北, 西南, 东南, go northwest, etc.

Raw direction parsing before movement command detection

Trilingual Display System

Text struct contains Chinese, Pinyin, and English fields

DisplayFormatter formats based on user preferences in config.json

Pinyin uses ASCII-only characters (no tone marks) to avoid Unicode parsing issues

Command Handling Priority

Quit commands checked FIRST to prevent movement parsing conflicts

Command categories: movement, item, save, info, quit

Spell-checking suggestions for common typos

Save System Design

Single save.go file (merged config and operations to eliminate duplication)

SaveData.Player as pointer (*models.Player) to match GameState.Player

Auto-save timer, quick-save/quick-load, multiple named slots

Inventory System

Dynamic capacity with backpack items that grant size bonuses

Equip/unequip system for weapons, armor, and backpacks

Inventory capacity checked before adding items

Data Files Structure
Required JSON Files (in data/ directory)
File	Purpose
config.json	Game settings, display preferences, combat balance
save_config.json	Save system configuration (slots, auto-save interval)
world.json	Rooms and locations with exits
items.json	Item definitions (weapons, armor, consumables, keys, backpacks)
enemies.json	Enemy stats, drops, biome affinity
biomes.json	Environmental descriptions and effects
Example World.json Structure
json
{
  "rooms": {
    "room_id": {
      "id": "room_id",
      "name": {"chinese": "...", "pinyin": "...", "english": "..."},
      "description": {"chinese": "...", "pinyin": "...", "english": "..."}
    }
  },
  "locations": {
    "location_id": {
      "id": "location_id",
      "room_id": "room_id",
      "biome_id": "biome_id",
      "name": {"chinese": "...", "pinyin": "...", "english": "..."},
      "exits": [
        {"direction": "north", "destination": "other_location", "requires_item_id": null}
      ],
      "enemy_ids": ["goblin"],
      "item_ids": ["sword"]
    }
  }
}
Building and Running
Prerequisites
Go 1.21 or higher

Build Commands
bash
# Navigate to project root
cd golden_chalice

# Download dependencies (uses only standard library)
go mod tidy

# Build the game
go build -o golden_chalice ./cmd/game/

# Run the game
./golden_chalice
Cross-Platform Builds
bash
# Windows
GOOS=windows GOARCH=amd64 go build -o golden_chalice.exe ./cmd/game/

# Linux
GOOS=linux GOARCH=amd64 go build -o golden_chalice ./cmd/game/

# macOS
GOOS=darwin GOARCH=amd64 go build -o golden_chalice ./cmd/game/
Command Reference
Category	Chinese	English
Movement	北, 往北, 往北走, 西北	go north, go northwest
Take	拿剑, 取剑	take sword
Inventory	背包, i	inventory, i
Status	状态	status
Look	看, 查看	look
Save	保存 slot1	save slot1
Load	加载 slot1	load slot1
Quit	退出, 退	quit, exit
Help	帮助	help

Future Expansion Possibilities
NPCs with dialogue trees

Quest system with multiple objectives

Skill/spell system

Crafting and item combination

Multiple endings based on choices

Day/night cycle affecting enemy spawns

Weather system with combat modifiers

Acknowledgements
Built with Go standard library only, demonstrating the power of pure Go for game development.