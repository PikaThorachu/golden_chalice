package game

import (
	"fmt"
	"strings"

	"golden_chalice/internal/errors"
	"golden_chalice/internal/models"
	"golden_chalice/internal/save"
)

// Command represents a parsed player command
type Command struct {
	Type     CommandType
	Args     []string
	RawInput string
}

// CommandType enumerates all possible commands
type CommandType int

const (
	CmdUnknown CommandType = iota
	CmdMove
	CmdTake
	CmdInventory
	CmdStatus
	CmdHelp
	CmdQuit
	CmdSave
	CmdLoad
	CmdListSaves
	CmdDeleteSave
	CmdLook
	CmdEquip
	CmdUnequip
	CmdUse
	CmdDrop
	CmdEquipBackpack
	CmdUnequipBackpack
	CmdInspectArea
	CmdInspectItem
	CmdOpen
	CmdTeleport
)

// CommandHandler processes commands and returns results
type CommandHandler struct {
	gameState      *GameState
	saveManager    *save.SaveManager
	inputValidator *InputValidator
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(gs *GameState, sm *save.SaveManager) *CommandHandler {
	return &CommandHandler{
		gameState:      gs,
		saveManager:    sm,
		inputValidator: NewInputValidator(),
	}
}

// formatOutput wraps a message with trilingual formatting
func (ch *CommandHandler) formatOutput(chinese, pinyin, english string) string {
	var parts []string

	if ch.gameState.Config.ShouldShowChinese() && chinese != "" {
		parts = append(parts, chinese)
	}

	if ch.gameState.Config.ShouldShowPinyin() && pinyin != "" {
		pinyinText := pinyin
		if ch.gameState.Config.ShouldShowChinese() && len(parts) > 0 {
			pinyinText = "(" + pinyinText + ")"
		}
		parts = append(parts, pinyinText)
	}

	if ch.gameState.Config.ShouldShowEnglish() && english != "" {
		englishText := english
		if len(parts) > 0 {
			englishText = "/ " + englishText
		}
		parts = append(parts, englishText)
	}

	if len(parts) == 0 {
		return chinese
	}

	return strings.Join(parts, " ")
}

// formatError returns a formatted error message
func (ch *CommandHandler) formatError(chinese, pinyin, english string) error {
	return fmt.Errorf(ch.formatOutput(chinese, pinyin, english))
}

// ProcessCommand parses and executes a player command
func (ch *CommandHandler) ProcessCommand(input string) (string, error) {
	result := ch.inputValidator.ValidateAndSanitize(input)
	if !result.IsValid {
		return "", ch.formatError(
			result.ErrorMsg,
			result.ErrorMsgPinyin,
			result.ErrorMsgEnglish,
		)
	}

	sanitized := result.Sanitized

	if ch.gameState.Logger != nil {
		ch.gameState.Logger.LogPlayerAction("command", map[string]interface{}{
			"input": sanitized,
		})
	}

	category := ch.inputValidator.GetCommandCategory(sanitized)

	// Check for teleport command first (testing)
	if strings.HasPrefix(sanitized, "tp ") || strings.HasPrefix(sanitized, "teleport ") || strings.HasPrefix(sanitized, "传送 ") {
		cmd := ch.parseCommand(sanitized)
		return ch.executeTeleport(cmd)
	}

	switch category {
	case "movement":
		moveResult := ch.inputValidator.ValidateMovementCommand(sanitized)
		if !moveResult.IsValid {
			return "", ch.formatError(
				moveResult.ErrorMsg,
				moveResult.ErrorMsgPinyin,
				moveResult.ErrorMsgEnglish,
			)
		}
		cmd := ch.parseCommand(sanitized)
		return ch.executeMove(cmd)

	case "item":
		cmdName, itemName, isValid, errMsg := ch.inputValidator.ValidateItemCommand(sanitized)
		if !isValid {
			return "", ch.formatError(errMsg, errMsg, errMsg)
		}
		switch cmdName {
		case "take":
			return ch.executeTake(Command{Type: CmdTake, Args: []string{itemName}})
		case "use":
			return ch.executeUse(Command{Type: CmdUse, Args: []string{itemName}})
		case "equip":
			return ch.executeEquip(Command{Type: CmdEquip, Args: []string{itemName}})
		case "drop":
			return ch.executeDrop(Command{Type: CmdDrop, Args: []string{itemName}})
		}

	case "save":
		cmdName, slotName, isValid, errMsg := ch.inputValidator.ValidateSaveCommand(sanitized)
		if !isValid {
			return "", ch.formatError(errMsg, errMsg, errMsg)
		}
		switch cmdName {
		case "save":
			return ch.executeSave(Command{Type: CmdSave, Args: []string{slotName}})
		case "load":
			return ch.executeLoad(Command{Type: CmdLoad, Args: []string{slotName}})
		case "delete":
			return ch.executeDeleteSave(Command{Type: CmdDeleteSave, Args: []string{slotName}})
		}

	case "inspect":
		cmd := ch.parseCommand(sanitized)
		switch cmd.Type {
		case CmdInspectArea:
			return ch.executeInspectArea()
		case CmdInspectItem:
			return ch.executeInspectItem(cmd)
		default:
			return "", ch.formatError(
				"无效的检查命令",
				"Wu xiao de jian cha ming ling",
				"Invalid inspect command",
			)
		}

	case "open":
		cmd := ch.parseCommand(sanitized)
		return ch.executeOpen(cmd)

	case "help", "inventory", "status", "look", "quit", "list_saves":
		cmd := ch.parseCommand(sanitized)
		switch cmd.Type {
		case CmdHelp:
			return ch.executeHelp()
		case CmdInventory:
			return ch.executeInventory()
		case CmdStatus:
			return ch.executeStatus()
		case CmdLook:
			return ch.executeLook()
		case CmdQuit:
			return ch.executeQuit()
		case CmdListSaves:
			return ch.executeListSaves()
		}

	default:
		suggestions := ch.inputValidator.SuggestCorrections(sanitized)
		if len(suggestions) > 0 {
			suggestionText := "您是否想输入: " + strings.Join(suggestions, ", ")
			return "", ch.formatError(suggestionText, suggestionText, suggestionText)
		}

		cmd := ch.parseCommand(sanitized)
		if cmd.Type != CmdUnknown {
			return ch.executeCommandByType(cmd)
		}
	}

	return "", ch.formatError(
		"未知命令。输入 '帮助' 查看可用命令",
		"Wei zhi ming ling. Shu ru 'bang zhu' cha kan ke yong ming ling",
		"Unknown command. Type 'help' to see available commands",
	)
}

// executeCommandByType is a helper method to execute command by type
func (ch *CommandHandler) executeCommandByType(cmd Command) (string, error) {
	switch cmd.Type {
	case CmdMove:
		return ch.executeMove(cmd)
	case CmdTake:
		return ch.executeTake(cmd)
	case CmdInventory:
		return ch.executeInventory()
	case CmdStatus:
		return ch.executeStatus()
	case CmdHelp:
		return ch.executeHelp()
	case CmdQuit:
		return ch.executeQuit()
	case CmdSave:
		return ch.executeSave(cmd)
	case CmdLoad:
		return ch.executeLoad(cmd)
	case CmdListSaves:
		return ch.executeListSaves()
	case CmdDeleteSave:
		return ch.executeDeleteSave(cmd)
	case CmdLook:
		return ch.executeLook()
	case CmdEquip:
		return ch.executeEquip(cmd)
	case CmdUnequip:
		return ch.executeUnequip(cmd)
	case CmdUse:
		return ch.executeUse(cmd)
	case CmdDrop:
		return ch.executeDrop(cmd)
	case CmdEquipBackpack:
		return ch.executeEquipBackpack(cmd)
	case CmdUnequipBackpack:
		return ch.executeUnequipBackpack()
	case CmdInspectArea:
		return ch.executeInspectArea()
	case CmdInspectItem:
		return ch.executeInspectItem(cmd)
	case CmdOpen:
		return ch.executeOpen(cmd)
	case CmdTeleport:
		return ch.executeTeleport(cmd)
	default:
		return "", ch.formatError(
			"未知命令",
			"Wei zhi ming ling",
			"Unknown command",
		)
	}
}

// parseCommand converts raw input into a structured Command
func (ch *CommandHandler) parseCommand(input string) Command {
	inputLower := strings.ToLower(input)

	inputLower = strings.ToLower(input)

	// ========== CHECK QUIT COMMANDS FIRST ==========
	if input == "退出" || input == "quit" || input == "exit" || input == "退" {
		return Command{Type: CmdQuit, RawInput: input}
	}

	// ========== INSPECT AREA COMMAND ==========
	if input == "检查周围" || input == "inspect area" || input == "查看周围" {
		return Command{Type: CmdInspectArea, RawInput: input}
	}

	// ========== INSPECT ITEM COMMAND ==========
	if strings.HasPrefix(input, "检查 ") || strings.HasPrefix(inputLower, "inspect ") {
		// Extract item name
		var itemName string
		if strings.HasPrefix(input, "检查 ") {
			itemName = strings.TrimPrefix(input, "检查 ")
		} else {
			itemName = strings.TrimPrefix(inputLower, "inspect ")
		}
		itemName = strings.TrimSpace(itemName)
		if itemName != "" {
			return Command{Type: CmdInspectItem, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// ========== OPEN COMMAND ==========
	if strings.HasPrefix(input, "打开 ") || strings.HasPrefix(inputLower, "open ") {
		var target string
		if strings.HasPrefix(input, "打开 ") {
			target = strings.TrimPrefix(input, "打开 ")
		} else {
			target = strings.TrimPrefix(inputLower, "open ")
		}
		target = strings.TrimSpace(target)
		if target != "" {
			return Command{Type: CmdOpen, Args: []string{target}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// ========== MOVEMENT COMMANDS ==========
	// Movement commands (Chinese format: 往北走, 往北去)
	if strings.HasPrefix(input, "往") {
		if dir, err := models.ParseDirection(input); err == nil {
			return Command{Type: CmdMove, Args: []string{dir.String()}, RawInput: input}
		}
	}

	// Movement commands (English format: go north, walk north)
	if strings.HasPrefix(inputLower, "go ") || strings.HasPrefix(inputLower, "walk ") {
		if dir, err := models.ParseDirection(input); err == nil {
			return Command{Type: CmdMove, Args: []string{dir.String()}, RawInput: input}
		}
	}

	// Raw Chinese direction (e.g., "北", "西北")
	if dir, err := models.ParseDirection(input); err == nil {
		return Command{Type: CmdMove, Args: []string{dir.String()}, RawInput: input}
	}

	// Teleport command for testing (only in debug mode)
	if strings.HasPrefix(input, "tp ") || strings.HasPrefix(input, "teleport ") || strings.HasPrefix(input, "传送 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			locationID := parts[1]
			return Command{Type: CmdTeleport, Args: []string{locationID}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// ========== TAKE COMMANDS ==========
	if strings.HasPrefix(input, "拿") || strings.HasPrefix(input, "取") {
		itemName := strings.TrimPrefix(input, "拿")
		itemName = strings.TrimPrefix(itemName, "取")
		itemName = strings.TrimSpace(itemName)
		if itemName != "" {
			return Command{Type: CmdTake, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// Take command (English)
	if strings.HasPrefix(inputLower, "take ") {
		itemName := strings.TrimPrefix(inputLower, "take ")
		itemName = strings.TrimSpace(itemName)
		if itemName != "" {
			return Command{Type: CmdTake, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// ========== INVENTORY COMMANDS ==========
	if input == "背包" || input == "i" || input == "inventory" {
		return Command{Type: CmdInventory, RawInput: input}
	}

	// ========== STATUS COMMANDS ==========
	if input == "状态" || input == "status" {
		return Command{Type: CmdStatus, RawInput: input}
	}

	// ========== HELP COMMANDS ==========
	if input == "帮助" || input == "help" {
		return Command{Type: CmdHelp, RawInput: input}
	}

	// ========== LOOK COMMANDS ==========
	if input == "看" || input == "查看" || input == "look" {
		return Command{Type: CmdLook, RawInput: input}
	}

	// ========== SAVE COMMANDS ==========
	if strings.HasPrefix(inputLower, "save ") || strings.HasPrefix(input, "保存 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			slotName := parts[1]
			return Command{Type: CmdSave, Args: []string{slotName}, RawInput: input}
		}
		return Command{Type: CmdSave, Args: []string{"autosave"}, RawInput: input}
	}

	// ========== LOAD COMMANDS ==========
	if strings.HasPrefix(inputLower, "load ") || strings.HasPrefix(input, "加载 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			slotName := parts[1]
			return Command{Type: CmdLoad, Args: []string{slotName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// ========== LIST SAVES COMMANDS ==========
	if input == "saves" || input == "存档列表" {
		return Command{Type: CmdListSaves, RawInput: input}
	}

	// ========== DELETE SAVE COMMANDS ==========
	if strings.HasPrefix(inputLower, "delete ") || strings.HasPrefix(input, "删除 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			slotName := parts[1]
			return Command{Type: CmdDeleteSave, Args: []string{slotName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// In parseCommand function, update the EQUIP section:

	// ========== EQUIP COMMANDS ==========
	// Chinese equip variations: 装备, 带上, 戴上, 佩上, 佩带, 穿着, 穿上
	if strings.HasPrefix(input, "装备") ||
		strings.HasPrefix(input, "带上") ||
		strings.HasPrefix(input, "戴上") ||
		strings.HasPrefix(input, "佩上") ||
		strings.HasPrefix(input, "佩带") ||
		strings.HasPrefix(input, "穿着") ||
		strings.HasPrefix(input, "穿上") {

		// Extract item name by removing the verb prefix
		itemName := input
		for _, prefix := range []string{"装备", "带上", "戴上", "佩上", "佩带", "穿着", "穿上"} {
			itemName = strings.TrimPrefix(itemName, prefix)
		}
		itemName = strings.TrimSpace(itemName)

		if itemName != "" {
			// Check if it's a backpack
			if itemName == "背包" || strings.Contains(itemName, "背包") {
				return Command{Type: CmdEquipBackpack, Args: []string{itemName}, RawInput: input}
			}
			return Command{Type: CmdEquip, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// ========== UNEQUIP COMMANDS ==========
	// Chinese unequip variations: 卸下, 取下, 脱下, 摘下, 解下
	if strings.HasPrefix(input, "卸下") ||
		strings.HasPrefix(input, "取下") ||
		strings.HasPrefix(input, "脱下") ||
		strings.HasPrefix(input, "摘下") ||
		strings.HasPrefix(input, "解下") {

		// Extract item name by removing the verb prefix
		slotOrItem := input
		for _, prefix := range []string{"卸下", "取下", "脱下", "摘下", "解下"} {
			slotOrItem = strings.TrimPrefix(slotOrItem, prefix)
		}
		slotOrItem = strings.TrimSpace(slotOrItem)

		if slotOrItem != "" {
			// Check if it's a backpack
			if slotOrItem == "背包" || slotOrItem == "backpack" {
				return Command{Type: CmdUnequipBackpack, RawInput: input}
			}
			return Command{Type: CmdUnequip, Args: []string{slotOrItem}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// ========== USE COMMANDS ==========
	if strings.HasPrefix(inputLower, "use ") || strings.HasPrefix(input, "使用 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			itemName := parts[1]
			return Command{Type: CmdUse, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	// ========== DROP COMMANDS ==========
	if strings.HasPrefix(inputLower, "drop ") || strings.HasPrefix(input, "丢弃 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			itemName := parts[1]
			return Command{Type: CmdDrop, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	return Command{Type: CmdUnknown, RawInput: input}
}

// executeMove handles movement commands
func (ch *CommandHandler) executeMove(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"往哪个方向走？",
			"Wang na ge fang xiang zou?",
			"Which direction?",
		)
	}

	direction := cmd.Args[0]

	var dir models.Direction
	switch direction {
	case "north":
		dir = models.North
	case "south":
		dir = models.South
	case "east":
		dir = models.East
	case "west":
		dir = models.West
	case "northwest":
		dir = models.Northwest
	case "northeast":
		dir = models.Northeast
	case "southwest":
		dir = models.Southwest
	case "southeast":
		dir = models.Southeast
	case "up":
		dir = models.Up
	case "down":
		dir = models.Down
	case "out":
		dir = models.Out
	case "in":
		dir = models.In
	default:
		return "", ch.formatError(
			"未知方向: "+direction,
			"Wei zhi fang xiang: "+direction,
			"Unknown direction: "+direction,
		)
	}

	err := ch.gameState.SafeMove(dir)
	if err != nil {
		if gameErr, ok := err.(*errors.GameError); ok {
			msg := gameErr.GetUserMessage(
				ch.gameState.Config.ShouldShowChinese(),
				ch.gameState.Config.ShouldShowPinyin(),
				ch.gameState.Config.ShouldShowEnglish(),
			)
			return "", fmt.Errorf("%s", msg)
		}
		return "", err
	}

	return ch.gameState.GetCurrentRoomDescription(), nil
}

// executeTeleport handles teleport command for testing
func (ch *CommandHandler) executeTeleport(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"传送去哪里？",
			"Chuan song qu na li?",
			"Teleport where?",
		)
	}

	locationID := cmd.Args[0]

	// Check if location exists
	_, err := ch.gameState.World.GetLocation(locationID)
	if err != nil {
		return "", ch.formatError(
			fmt.Sprintf("位置 '%s' 不存在", locationID),
			fmt.Sprintf("Wei zhi '%s' bu cun zai", locationID),
			fmt.Sprintf("Location '%s' does not exist", locationID),
		)
	}

	// Store old location for logging
	oldLocationID := ch.gameState.Player.CurrentLocationID

	// Teleport player
	ch.gameState.Player.CurrentLocationID = locationID
	ch.gameState.LastLocationID = oldLocationID

	// Update room tracking
	ch.gameState.updateCurrentRoom()

	// Log teleport
	if ch.gameState.Logger != nil {
		ch.gameState.Logger.LogPlayerAction("teleport", map[string]interface{}{
			"from": oldLocationID,
			"to":   locationID,
		})
	}

	// Return new room description
	result := ch.formatOutput(
		fmt.Sprintf("传送到了: %s", locationID),
		fmt.Sprintf("Chuan song dao le: %s", locationID),
		fmt.Sprintf("Teleported to: %s", locationID),
	) + "\n" + ch.gameState.GetCurrentRoomDescription()

	return result, nil
}

// executeTake handles item pickup commands
func (ch *CommandHandler) executeTake(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"拿什么？",
			"Na shen me?",
			"Take what?",
		)
	}

	itemName := cmd.Args[0]

	itemsAtLocation := ch.gameState.GetItemsAtCurrentLocation()
	var matchedItemID string
	var matchedItem models.Item

	for _, itemID := range itemsAtLocation {
		if item, exists := ch.gameState.Items[itemID]; exists {
			if item.Name.Chinese == itemName || strings.EqualFold(item.Name.English, itemName) {
				matchedItemID = itemID
				matchedItem = item
				break
			}
		}
	}

	if matchedItemID == "" {
		return "", ch.formatError(
			"这里没有 "+itemName,
			"Zhe li mei you "+itemName,
			"There is no "+itemName+" here",
		)
	}

	err := ch.gameState.SafeTakeItem(matchedItemID)
	if err != nil {
		if gameErr, ok := err.(*errors.GameError); ok {
			return "", fmt.Errorf(gameErr.GetUserMessage(
				ch.gameState.Config.ShouldShowChinese(),
				ch.gameState.Config.ShouldShowPinyin(),
				ch.gameState.Config.ShouldShowEnglish(),
			))
		}
		return "", err
	}

	return ch.formatOutput(
		"你拿起了 "+matchedItem.Name.Chinese,
		"Ni na qi le "+matchedItem.Name.Pinyin,
		"You picked up: "+matchedItem.Name.English,
	), nil
}

// executeInventory displays player inventory
func (ch *CommandHandler) executeInventory() (string, error) {
	return ch.gameState.GetInventoryDisplay(), nil
}

// executeInspectArea inspects all nearby locations
func (ch *CommandHandler) executeInspectArea() (string, error) {
	nearbyItems, err := ch.gameState.GetNearbyItems()
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString(ch.formatOutput(
		"=== 周围环境检查 ===",
		"=== Zhou Wei Huan Jing Jian Cha ===",
		"=== Environment Inspection ===",
	))
	result.WriteString("\n")

	// Group items by location
	locationItems := make(map[string][]struct {
		LocationName models.Text
		Item         models.Item
		IsContainer  bool
		IsDrop       bool
	})

	for _, item := range nearbyItems {
		locationItems[item.LocationID] = append(locationItems[item.LocationID], struct {
			LocationName models.Text
			Item         models.Item
			IsContainer  bool
			IsDrop       bool
		}{
			LocationName: item.LocationName,
			Item:         item.Item,
			IsContainer:  item.IsContainer,
			IsDrop:       item.IsDrop,
		})
	}

	for locID, items := range locationItems {
		loc, _ := ch.gameState.World.GetLocation(locID)
		result.WriteString(ch.formatOutput(
			fmt.Sprintf("\n%s:", loc.Name.Chinese),
			fmt.Sprintf("%s:", loc.Name.Pinyin),
			fmt.Sprintf("%s:", loc.Name.English),
		))
		result.WriteString("\n")

		for _, item := range items {
			itemType := "物品"
			itemTypePinyin := "wu pin"
			itemTypeEnglish := "item"

			if item.IsContainer {
				itemType = "容器"
				itemTypePinyin = "rong qi"
				itemTypeEnglish = "container"
			}

			result.WriteString(fmt.Sprintf("  • %s (%s)\n",
				ch.formatOutput(
					item.Item.Name.Chinese,
					item.Item.Name.Pinyin,
					item.Item.Name.English,
				),
				ch.formatOutput(itemType, itemTypePinyin, itemTypeEnglish),
			))
		}
	}

	if len(nearbyItems) == 0 {
		result.WriteString(ch.formatOutput(
			"附近没有发现任何物品或容器。",
			"Fu jin mei you fa xian ren he wu pin huo rong qi.",
			"No items or containers found nearby.",
		))
	}

	return result.String(), nil
}

// executeInspectItem inspects a specific item for locks, traps, spells
func (ch *CommandHandler) executeInspectItem(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"检查什么物品？",
			"Jian cha shen me wu pin?",
			"What item to inspect?",
		)
	}

	itemName := cmd.Args[0]

	// First, check items in current location
	currentItems := ch.gameState.GetItemsAtCurrentLocation()
	var targetItemID string
	var targetItem models.Item
	var foundInCurrentLocation bool

	for _, itemID := range currentItems {
		if item, exists := ch.gameState.Items[itemID]; exists {
			// Check by Chinese name, English name, or ID
			if item.Name.Chinese == itemName ||
				strings.EqualFold(item.Name.English, itemName) ||
				item.ID == itemName {
				targetItemID = itemID
				targetItem = item
				foundInCurrentLocation = true
				break
			}
		}
	}

	// If not found in current location, check nearby locations
	if !foundInCurrentLocation {
		currentLoc, err := ch.gameState.GetCurrentRoom()
		if err != nil {
			return "", err
		}

		nearbyItems, err := ch.gameState.World.GetItemsInNearbyLocations(
			currentLoc.ID,
			ch.gameState.TakenItems,
			ch.gameState.PendingDrops,
			ch.gameState.Items,
		)
		if err != nil {
			return "", err
		}

		for _, itemInfo := range nearbyItems {
			if itemInfo.Item.Name.Chinese == itemName ||
				strings.EqualFold(itemInfo.Item.Name.English, itemName) ||
				itemInfo.Item.ID == itemName {
				targetItem = itemInfo.Item
				targetItemID = targetItem.ID
				foundInCurrentLocation = false
				break
			}
		}

		if targetItemID == "" {
			return "", ch.formatError(
				fmt.Sprintf("在附近找不到物品: %s", itemName),
				fmt.Sprintf("Zai fu jin zhao bu dao wu pin: %s", itemName),
				fmt.Sprintf("Cannot find item nearby: %s", itemName),
			)
		}
	}

	// Build inspection result
	var result strings.Builder
	result.WriteString(ch.formatOutput(
		fmt.Sprintf("=== 检查 %s ===", targetItem.Name.Chinese),
		fmt.Sprintf("=== Jian Cha %s ===", targetItem.Name.Pinyin),
		fmt.Sprintf("=== Inspecting %s ===", targetItem.Name.English),
	))
	result.WriteString("\n")

	// Check if it's a container
	if targetItem.IsContainer() {
		result.WriteString(ch.formatOutput(
			fmt.Sprintf("  这是一个容器，可以容纳 %d 件物品。", targetItem.GetCapacity()),
			fmt.Sprintf("  Zhe shi yi ge rong qi, ke yi rong na %d jian wu pin.", targetItem.GetCapacity()),
			fmt.Sprintf("  This is a container that can hold %d items.", targetItem.GetCapacity()),
		))
		result.WriteString("\n")

		// Show contents if any
		if len(targetItem.Inventory) > 0 {
			result.WriteString(ch.formatOutput(
				"  里面包含:",
				"  Li mian bao han:",
				"  Contains:",
			))
			result.WriteString("\n")
			for _, containedID := range targetItem.Inventory {
				if containedItem, exists := ch.gameState.Items[containedID]; exists {
					result.WriteString(fmt.Sprintf("    • %s\n",
						ch.formatOutput(
							containedItem.Name.Chinese,
							containedItem.Name.Pinyin,
							containedItem.Name.English,
						)))
				}
			}
		}
	}

	// Check for locks
	if targetItem.Properties.OpensDoorID != nil && *targetItem.Properties.OpensDoorID != "" {
		result.WriteString(ch.formatOutput(
			"  这是一个钥匙，可以打开特定的门。",
			"  Zhe shi yi ge yao chi, ke yi da kai te ding de men.",
			"  This is a key that can open specific doors.",
		))
		result.WriteString("\n")
	}

	// Check for traps
	if targetItem.GetTrapEnemyID() != "" {
		result.WriteString(ch.formatOutput(
			"  ⚠️ 这个物品带有陷阱！",
			"  ⚠️ Zhe ge wu pin dai you xian jing!",
			"  ⚠️ This item is trapped!",
		))
		result.WriteString("\n")
	}

	// Check for spells/enchantments
	if targetItem.IsWeapon() && targetItem.GetDamageBonus() > 15 {
		result.WriteString(ch.formatOutput(
			"  这把武器散发着魔法光芒。",
			"  Zhe ba wu qi san fa zhe mo fa guang mang.",
			"  This weapon radiates magical energy.",
		))
		result.WriteString("\n")
	}

	// If nothing special found
	if targetItem.GetTrapEnemyID() == "" &&
		targetItem.Properties.OpensDoorID == nil &&
		!targetItem.IsContainer() &&
		!(targetItem.IsWeapon() && targetItem.GetDamageBonus() > 15) {
		result.WriteString(ch.formatOutput(
			"  这个物品看起来很普通，没有什么特别之处。",
			"  Zhe ge wu pin kan qi lai hen pu tong, mei you shen me te bie zhi chu.",
			"  This item looks ordinary with nothing special about it.",
		))
	}

	return result.String(), nil
}

// executeOpen opens doors or containers
func (ch *CommandHandler) executeOpen(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"打开什么？",
			"Da kai shen me?",
			"Open what?",
		)
	}

	target := cmd.Args[0]

	// Check if target is a door (exit with required item)
	currentLoc, err := ch.gameState.GetCurrentRoom()
	if err != nil {
		return "", err
	}

	// Check if target is an exit direction
	var dir models.Direction
	switch target {
	case "北", "北门", "north":
		dir = models.North
	case "南", "南门", "south":
		dir = models.South
	case "东", "东门", "east":
		dir = models.East
	case "西", "西门", "west":
		dir = models.West
	case "西北", "西北门", "northwest":
		dir = models.Northwest
	case "东北", "东北门", "northeast":
		dir = models.Northeast
	case "西南", "西南门", "southwest":
		dir = models.Southwest
	case "东南", "东南门", "southeast":
		dir = models.Southeast
	default:
		// Check if target is a container item
		return ch.openContainer(target)
	}

	// Try to open a door
	exit, err := ch.gameState.World.GetExit(currentLoc.ID, dir)
	if err != nil {
		return "", ch.formatError(
			"那里没有门。",
			"Na li mei you men.",
			"There is no door there.",
		)
	}

	if exit.RequiredItem == nil {
		return "", ch.formatError(
			"这门没有锁。",
			"Zhe men mei you suo.",
			"This door is not locked.",
		)
	}

	// Check if player has the key
	requiredItem := *exit.RequiredItem
	if !ch.gameState.Player.HasItem(requiredItem) {
		return "", ch.formatError(
			fmt.Sprintf("门被锁住了。需要 %s 才能打开。", requiredItem),
			fmt.Sprintf("Men bei suo zhu le. Xu yao %s cai neng da kai.", requiredItem),
			fmt.Sprintf("The door is locked. Need %s to open it.", requiredItem),
		)
	}

	// Remove the key from inventory (optional - key might be consumed)
	ch.gameState.Player.RemoveItem(requiredItem)

	// Mark the exit as unlocked by setting RequiredItem to nil
	// Note: This modifies the world - consider using a separate tracking map
	for i, e := range currentLoc.Exits {
		if e.Direction == dir {
			currentLoc.Exits[i].RequiredItem = nil
			ch.gameState.World.Locations[currentLoc.ID] = currentLoc
			break
		}
	}

	return ch.formatOutput(
		fmt.Sprintf("你用 %s 打开了门。", requiredItem),
		fmt.Sprintf("Ni yong %s da kai le men.", requiredItem),
		fmt.Sprintf("You opened the door with %s.", requiredItem),
	), nil
}

// openContainer handles opening containers like chests
func (ch *CommandHandler) openContainer(containerName string) (string, error) {
	// Find container in current location
	currentItems := ch.gameState.GetItemsAtCurrentLocation()
	var containerID string
	var container models.Item

	for _, itemID := range currentItems {
		if item, exists := ch.gameState.Items[itemID]; exists {
			if (item.Name.Chinese == containerName || strings.EqualFold(item.Name.English, containerName)) && item.IsContainer() {
				containerID = itemID
				container = item
				break
			}
		}
	}

	if containerID == "" {
		return "", ch.formatError(
			fmt.Sprintf("附近没有 %s 容器。", containerName),
			fmt.Sprintf("Fu jin mei you %s rong qi.", containerName),
			fmt.Sprintf("No container named %s nearby.", containerName),
		)
	}

	// Display container contents
	var result strings.Builder
	result.WriteString(ch.formatOutput(
		fmt.Sprintf("=== 打开 %s ===", container.Name.Chinese),
		fmt.Sprintf("=== Da Kai %s ===", container.Name.Pinyin),
		fmt.Sprintf("=== Opening %s ===", container.Name.English),
	))
	result.WriteString("\n")

	if len(container.Inventory) == 0 {
		result.WriteString(ch.formatOutput(
			"容器是空的。",
			"Rong qi shi kong de.",
			"The container is empty.",
		))
	} else {
		result.WriteString(ch.formatOutput(
			"里面发现:",
			"Li mian fa xian:",
			"Inside you find:",
		))
		result.WriteString("\n")

		for _, itemID := range container.Inventory {
			if item, exists := ch.gameState.Items[itemID]; exists {
				result.WriteString(fmt.Sprintf("  • %s\n",
					ch.formatOutput(
						item.Name.Chinese,
						item.Name.Pinyin,
						item.Name.English,
					),
				))
			}
		}

		// Option to take items from container
		result.WriteString(ch.formatOutput(
			"\n使用 '拿 <物品>' 从容器中取出物品。",
			"\nShi yong 'na <wu pin>' cong rong qi zhong qu chu wu pin.",
			"\nUse 'take <item>' to take items from the container.",
		))
	}

	return result.String(), nil
}

// handleTrapInspect handles combat triggered by inspection
func (ch *CommandHandler) handleTrapInspect(item models.Item, _ *models.Enemy) (string, error) {
	// Remove the trap item from location (it was a mimic/chest)
	// Add the enemy to current location
	// Then trigger combat
	return ch.formatOutput(
		item.GetInspectMessage(),
		item.GetInspectMessage(),
		item.GetInspectMessage(),
	), nil
}

// executeStatus displays player status
func (ch *CommandHandler) executeStatus() (string, error) {
	return ch.gameState.GetPlayerStatus(), nil
}

// executeHelp displays help text
// executeHelp displays help text
func (ch *CommandHandler) executeHelp() (string, error) {
	var result strings.Builder

	// Header
	result.WriteString("\n╔═══════════════════════════════════════════════════════════╗")

	// Chinese header
	if ch.gameState.Config.ShouldShowChinese() {
		result.WriteString("\n║                      游戏命令帮助                         ║")
	}

	// Pinyin header
	if ch.gameState.Config.ShouldShowPinyin() {
		result.WriteString("\n║                    You Xi Ming Ling Bang Zhu              ║")
	}

	// English header
	if ch.gameState.Config.ShouldShowEnglish() {
		result.WriteString("\n║                      Game Commands                        ║")
	}

	result.WriteString("\n╠═══════════════════════════════════════════════════════════╣")

	// ==================== MOVEMENT ====================
	// Chinese movement
	if ch.gameState.Config.ShouldShowChinese() {
		result.WriteString("\n║ 移动：                                                     ║")
		result.WriteString("\n║     往<北/南/东/西/西北/东北/西南/东南>走                  ║")
		result.WriteString("\n║     往<出>走                                               ║")
	}

	// Pinyin movement
	if ch.gameState.Config.ShouldShowPinyin() {
		result.WriteString("\n║ Yi Dong:                                                  ║")
		result.WriteString("\n║     Wang <bei/nan/dong/xi/xi bei/dong bei/xi nan/dong bei> zou ║")
		result.WriteString("\n║     Wang <chu> zou                                        ║")
	}

	// English movement
	if ch.gameState.Config.ShouldShowEnglish() {
		result.WriteString("\n║ Movement:                                                 ║")
		result.WriteString("\n║     Go <north/south/east/west/northwest/northeast/southwest/southeast> ║")
		result.WriteString("\n║     Go <out>                                              ║")
	}

	result.WriteString("\n╠═══════════════════════════════════════════════════════════╣")

	// ==================== ITEM OPERATIONS ====================
	// Chinese item operations
	if ch.gameState.Config.ShouldShowChinese() {
		result.WriteString("\n║ 物品操作：                                                 ║")
		result.WriteString("\n║    拿取: 拿<物品名>                                        ║")
		result.WriteString("\n║    使用: 使用<物品名>                                      ║")
		result.WriteString("\n║    装备: 装备<物品名>                                      ║")
		result.WriteString("\n║    卸下: 卸下<武器/护甲>                                   ║")
		result.WriteString("\n║    丢弃: 丢弃<物品名>                                      ║")
		result.WriteString("\n║    检查: 检查<物品名>                                      ║")
	}

	// Pinyin item operations
	if ch.gameState.Config.ShouldShowPinyin() {
		result.WriteString("\n║ Wu Pin Cao Zuo:                                           ║")
		result.WriteString("\n║     Na Qu: Na <wu pin ming>                               ║")
		result.WriteString("\n║     Shi Yong: Shi Yong <wu pin ming>                     ║")
		result.WriteString("\n║     Zhuang Bei: Zhuang Bei <wu pin ming>                 ║")
		result.WriteString("\n║     Xie Xia: Xie Xia <wu qi / hu jia>                    ║")
		result.WriteString("\n║     Diu Qi: Diu Qi <wu pin ming>                         ║")
		result.WriteString("\n║     Jian Cha: Jian Cha <wu pin ming>                     ║")
	}

	// English item operations
	if ch.gameState.Config.ShouldShowEnglish() {
		result.WriteString("\n║ Item Operations:                                          ║")
		result.WriteString("\n║     Take: take <item name>                               ║")
		result.WriteString("\n║     Use: use <item name>                                 ║")
		result.WriteString("\n║     Equip: equip <item name>                             ║")
		result.WriteString("\n║     Unequip: unequip <weapon/armor>                      ║")
		result.WriteString("\n║     Drop: drop <item name>                               ║")
		result.WriteString("\n║     Inspect: inspect <item name>                         ║")
	}

	result.WriteString("\n╠═══════════════════════════════════════════════════════════╣")

	// ==================== INFORMATION ====================
	// Chinese information
	if ch.gameState.Config.ShouldShowChinese() {
		result.WriteString("\n║ 信息查看：                                                 ║")
		result.WriteString("\n║    背包: 背包 或 i                                         ║")
		result.WriteString("\n║    状态: 状态                                              ║")
		result.WriteString("\n║    查看: 看 或 查看                                        ║")
		result.WriteString("\n║    帮助: 帮助                                              ║")
	}

	// Pinyin information
	if ch.gameState.Config.ShouldShowPinyin() {
		result.WriteString("\n║ Xin Xi Cha Kan:                                           ║")
		result.WriteString("\n║     Bei Bao: Bei Bao huo i                                ║")
		result.WriteString("\n║     Zhuang Tai: Zhuang Tai                                ║")
		result.WriteString("\n║     Cha Kan: Kan huo Cha Kan                              ║")
		result.WriteString("\n║     Bang Zhu: Bang Zhu                                    ║")
	}

	// English information
	if ch.gameState.Config.ShouldShowEnglish() {
		result.WriteString("\n║ Information:                                              ║")
		result.WriteString("\n║     Inventory: inventory or i                            ║")
		result.WriteString("\n║     Status: status                                       ║")
		result.WriteString("\n║     Look: look                                           ║")
		result.WriteString("\n║     Help: help                                           ║")
	}

	result.WriteString("\n╠═══════════════════════════════════════════════════════════╣")

	// ==================== SAVE/LOAD ====================
	// Chinese save/load
	if ch.gameState.Config.ShouldShowChinese() {
		result.WriteString("\n║ 存档管理：                                                 ║")
		result.WriteString("\n║    保存: save <名称> 或 保存 <名称>                        ║")
		result.WriteString("\n║    加载: load <名称> 或 加载 <名称>                        ║")
		result.WriteString("\n║    列表: saves 或 存档列表                                 ║")
		result.WriteString("\n║    删除: delete <名称> 或 删除 <名称>                      ║")
	}

	// Pinyin save/load
	if ch.gameState.Config.ShouldShowPinyin() {
		result.WriteString("\n║ Cun Dang Guan Li:                                        ║")
		result.WriteString("\n║     Bao Cun: save <ming cheng> huo Bao Cun <ming cheng>  ║")
		result.WriteString("\n║     Jia Zai: load <ming cheng> huo Jia Zai <ming cheng>  ║")
		result.WriteString("\n║     Lie Biao: saves huo Cun Dang Lie Biao                ║")
		result.WriteString("\n║     Shan Chu: delete <ming cheng> huo Shan Chu <ming cheng>║")
	}

	// English save/load
	if ch.gameState.Config.ShouldShowEnglish() {
		result.WriteString("\n║ Save/Load:                                                ║")
		result.WriteString("\n║     Save: save <name>                                    ║")
		result.WriteString("\n║     Load: load <name>                                    ║")
		result.WriteString("\n║     List: saves                                          ║")
		result.WriteString("\n║     Delete: delete <name>                                ║")
	}

	result.WriteString("\n╠═══════════════════════════════════════════════════════════╣")

	// ==================== OTHER ====================
	// Chinese other
	if ch.gameState.Config.ShouldShowChinese() {
		result.WriteString("\n║ 其他：                                                     ║")
		result.WriteString("\n║    退出: quit 或 退出                                      ║")
		result.WriteString("\n║    传送: tp <位置> 或 传送 <位置> (测试用)                  ║")
	}

	// Pinyin other
	if ch.gameState.Config.ShouldShowPinyin() {
		result.WriteString("\n║ Qi Ta:                                                    ║")
		result.WriteString("\n║     Tui Chu: quit huo Tui Chu                            ║")
		result.WriteString("\n║     Chuan Song: tp <wei zhi> huo Chuan Song <wei zhi>    ║")
	}

	// English other
	if ch.gameState.Config.ShouldShowEnglish() {
		result.WriteString("\n║ Other:                                                    ║")
		result.WriteString("\n║     Quit: quit                                           ║")
		result.WriteString("\n║     Teleport: tp <location> (testing)                    ║")
	}

	result.WriteString("\n╚═══════════════════════════════════════════════════════════╝")

	return result.String(), nil
}

// executeQuit handles quit command - now just sets the flag
func (ch *CommandHandler) executeQuit() (string, error) {
	ch.gameState.GameOver = true
	return ch.formatOutput(
		"感谢游玩！再见！",
		"Gan xie you wan! Zai jian!",
		"Thanks for playing! Farewell!",
	), nil
}

// executeSave handles save command
func (ch *CommandHandler) executeSave(cmd Command) (string, error) {
	if ch.saveManager == nil {
		return "", ch.formatError(
			"保存系统未初始化",
			"Bao cun xi tong wei chu shi hua",
			"Save system not initialized",
		)
	}

	slotName := "autosave"
	if len(cmd.Args) > 0 {
		slotName = cmd.Args[0]
	}

	_, err := ch.saveManager.CreateSave(
		ch.gameState.Player,
		ch.gameState.DefeatedEnemies,
		ch.gameState.TakenItems,
		ch.gameState.PendingDrops,
		ch.gameState.Config.GameVersion,
		ch.gameState.Player.Name,
	)
	if err != nil {
		return "", ch.formatError(
			fmt.Sprintf("创建存档失败: %v", err),
			fmt.Sprintf("Chuang jian cun dang shi bai: %v", err),
			fmt.Sprintf("Failed to create save: %v", err),
		)
	}

	err = ch.saveManager.SaveToFile(slotName)
	if err != nil {
		return "", ch.formatError(
			fmt.Sprintf("保存失败: %v", err),
			fmt.Sprintf("Bao cun shi bai: %v", err),
			fmt.Sprintf("Save failed: %v", err),
		)
	}

	return ch.formatOutput(
		fmt.Sprintf("游戏已保存到: %s", slotName),
		fmt.Sprintf("You xi yi bao cun dao: %s", slotName),
		fmt.Sprintf("Game saved to: %s", slotName),
	), nil
}

// executeLoad handles load command
func (ch *CommandHandler) executeLoad(cmd Command) (string, error) {
	if ch.saveManager == nil {
		return "", ch.formatError(
			"保存系统未初始化",
			"Bao cun xi tong wei chu shi hua",
			"Save system not initialized",
		)
	}

	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"请指定要加载的存档名称",
			"Qing zhi ding yao jia zai de cun dang ming cheng",
			"Please specify a save slot to load",
		)
	}

	slotName := cmd.Args[0]

	saveData, err := ch.saveManager.LoadFromFile(slotName)
	if err != nil {
		return "", ch.formatError(
			fmt.Sprintf("加载失败: %v", err),
			fmt.Sprintf("Jia zai shi bai: %v", err),
			fmt.Sprintf("Load failed: %v", err),
		)
	}

	// Restore game state - saveData.Player is already *models.Player
	ch.gameState.Player = saveData.Player
	ch.gameState.DefeatedEnemies = saveData.DefeatedEnemies
	ch.gameState.TakenItems = saveData.TakenItems
	ch.gameState.PendingDrops = saveData.PendingDrops
	ch.gameState.GameOver = false
	ch.gameState.GameWon = false

	// Reset room tracking
	ch.gameState.CurrentRoomID = nil
	ch.gameState.LastLocationID = ""

	result := ch.formatOutput(
		fmt.Sprintf("已加载存档: %s", slotName),
		fmt.Sprintf("Yi jia zai cun dang: %s", slotName),
		fmt.Sprintf("Loaded save: %s", slotName),
	) + "\n" + ch.gameState.GetCurrentRoomDescription()

	return result, nil
}

// executeListSaves lists all available saves
func (ch *CommandHandler) executeListSaves() (string, error) {
	if ch.saveManager == nil {
		return "", ch.formatError(
			"保存系统未初始化",
			"Bao cun xi tong wei chu shi hua",
			"Save system not initialized",
		)
	}

	saves, err := ch.saveManager.ListSaves()
	if err != nil {
		return "", ch.formatError(
			fmt.Sprintf("无法列出存档: %v", err),
			fmt.Sprintf("Wu fa lie chu cun dang: %v", err),
			fmt.Sprintf("Failed to list saves: %v", err),
		)
	}

	if len(saves) == 0 {
		return ch.formatOutput(
			"没有找到任何存档",
			"Mei you zhao dao ren he cun dang",
			"No saves found",
		), nil
	}

	var result strings.Builder
	result.WriteString(ch.formatOutput("=== 存档列表 ===", "=== Cun dang lie biao ===", "=== Save List ===") + "\n")
	for _, s := range saves {
		result.WriteString(fmt.Sprintf("  %s - %s - %s\n",
			s.DisplayName,
			s.PlayerName,
			s.SaveTime.Format("2006-01-02 15:04:05")))
	}
	result.WriteString(ch.formatOutput(
		fmt.Sprintf("\n共 %d 个存档", len(saves)),
		fmt.Sprintf("Gong %d ge cun dang", len(saves)),
		fmt.Sprintf("Total: %d saves", len(saves)),
	))

	return result.String(), nil
}

// executeDeleteSave deletes a save file
func (ch *CommandHandler) executeDeleteSave(cmd Command) (string, error) {
	if ch.saveManager == nil {
		return "", ch.formatError(
			"保存系统未初始化",
			"Bao cun xi tong wei chu shi hua",
			"Save system not initialized",
		)
	}

	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"请指定要删除的存档名称",
			"Qing zhi ding yao shan chu de cun dang ming cheng",
			"Please specify a save slot to delete",
		)
	}

	slotName := cmd.Args[0]

	err := ch.saveManager.DeleteSave(slotName)
	if err != nil {
		return "", ch.formatError(
			fmt.Sprintf("删除失败: %v", err),
			fmt.Sprintf("Shan chu shi bai: %v", err),
			fmt.Sprintf("Delete failed: %v", err),
		)
	}

	return ch.formatOutput(
		fmt.Sprintf("已删除存档: %s", slotName),
		fmt.Sprintf("Yi shan chu cun dang: %s", slotName),
		fmt.Sprintf("Deleted save: %s", slotName),
	), nil
}

// executeLook handles look command
func (ch *CommandHandler) executeLook() (string, error) {
	return ch.gameState.GetLookDescription(), nil
}

// executeEquip handles equipment commands
func (ch *CommandHandler) executeEquip(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"装备什么？",
			"Zhuang bei shen me?",
			"Equip what?",
		)
	}

	itemName := cmd.Args[0]

	var itemID string
	var foundItem models.Item
	for _, id := range ch.gameState.Player.Inventory {
		if item, exists := ch.gameState.Items[id]; exists {
			if item.Name.Chinese == itemName || strings.EqualFold(item.Name.English, itemName) {
				itemID = id
				foundItem = item
				break
			}
		}
	}

	if itemID == "" {
		return "", ch.formatError(
			"你没有 "+itemName,
			"Ni mei you "+itemName,
			"You don't have "+itemName,
		)
	}

	// Determine appropriate verb based on item type
	var equipVerb string
	var equipVerbPinyin string
	var equipVerbEnglish string

	if foundItem.IsWeapon() {
		equipVerb = "佩上了"
		equipVerbPinyin = "Pei shang le"
		equipVerbEnglish = "equipped"
	} else if foundItem.IsArmor() {
		equipVerb = "穿上了"
		equipVerbPinyin = "Chuan shang le"
		equipVerbEnglish = "put on"
	} else if foundItem.IsBackpack() {
		equipVerb = "带上了"
		equipVerbPinyin = "Dai shang le"
		equipVerbEnglish = "equipped"
	} else if strings.Contains(foundItem.Name.Chinese, "戒指") || strings.Contains(foundItem.Name.English, "ring") {
		equipVerb = "戴上了"
		equipVerbPinyin = "Dai shang le"
		equipVerbEnglish = "put on"
	} else if strings.Contains(foundItem.Name.Chinese, "项链") || strings.Contains(foundItem.Name.English, "necklace") {
		equipVerb = "戴上了"
		equipVerbPinyin = "Dai shang le"
		equipVerbEnglish = "put on"
	} else {
		equipVerb = "装备了"
		equipVerbPinyin = "Zhuang bei le"
		equipVerbEnglish = "equipped"
	}

	if foundItem.IsWeapon() {
		ch.gameState.Player.EquipWeapon(itemID)
		return ch.formatOutput(
			equipVerb+" "+foundItem.Name.Chinese,
			equipVerbPinyin+" "+foundItem.Name.Pinyin,
			equipVerbEnglish+" "+foundItem.Name.English,
		), nil
	}

	if foundItem.IsArmor() {
		ch.gameState.Player.EquipArmor(itemID)
		return ch.formatOutput(
			equipVerb+" "+foundItem.Name.Chinese,
			equipVerbPinyin+" "+foundItem.Name.Pinyin,
			equipVerbEnglish+" "+foundItem.Name.English,
		), nil
	}

	return "", ch.formatError(
		foundItem.Name.Chinese+" 无法装备",
		foundItem.Name.Pinyin+" wu fa zhuang bei",
		foundItem.Name.English+" cannot be equipped",
	)
}

// executeUnequip handles unequip commands
func (ch *CommandHandler) executeUnequip(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"卸下什么？（武器/护甲）",
			"Xie xia shen me? (wu qi / hu jia)",
			"Unequip what? (weapon/armor)",
		)
	}

	slot := cmd.Args[0]

	// Determine appropriate verb based on item type
	var unequipVerb string
	var unequipVerbPinyin string
	var unequipVerbEnglish string

	if slot == "武器" || slot == "weapon" {
		if ch.gameState.Player.HasEquippedWeapon() {
			weaponID := ch.gameState.Player.GetEquippedWeaponID()
			weapon, exists := ch.gameState.Items[weaponID]

			// Choose verb based on weapon type
			if strings.Contains(weapon.Name.Chinese, "剑") {
				unequipVerb = "解下了"
				unequipVerbPinyin = "Jie xia le"
				unequipVerbEnglish = "unstrapped"
			} else {
				unequipVerb = "卸下了"
				unequipVerbPinyin = "Xie xia le"
				unequipVerbEnglish = "unequipped"
			}

			ch.gameState.Player.UnequipWeapon()

			if exists {
				return ch.formatOutput(
					unequipVerb+" "+weapon.Name.Chinese,
					unequipVerbPinyin+" "+weapon.Name.Pinyin,
					unequipVerbEnglish+" "+weapon.Name.English,
				), nil
			}
			return ch.formatOutput(
				"已卸下武器",
				"Yi xie xia wu qi",
				"Weapon unequipped",
			), nil
		}
		return "", ch.formatError(
			"没有装备武器",
			"Mei you zhuang bei wu qi",
			"No weapon equipped",
		)
	}

	if slot == "护甲" || slot == "armor" {
		if ch.gameState.Player.HasEquippedArmor() {
			armorID := ch.gameState.Player.GetEquippedArmorID()
			armor, exists := ch.gameState.Items[armorID]

			unequipVerb = "脱下了"
			unequipVerbPinyin = "Tuo xia le"
			unequipVerbEnglish = "took off"

			ch.gameState.Player.UnequipArmor()

			if exists {
				return ch.formatOutput(
					unequipVerb+" "+armor.Name.Chinese,
					unequipVerbPinyin+" "+armor.Name.Pinyin,
					unequipVerbEnglish+" "+armor.Name.English,
				), nil
			}
			return ch.formatOutput(
				"已卸下护甲",
				"Yi xie xia hu jia",
				"Armor unequipped",
			), nil
		}
		return "", ch.formatError(
			"没有装备护甲",
			"Mei you zhuang bei hu jia",
			"No armor equipped",
		)
	}

	return "", ch.formatError(
		"未知装备槽: "+slot,
		"Wei zhi zhuang bei cao: "+slot,
		"Unknown equipment slot: "+slot,
	)
}

// executeUse handles using items (potions, etc.)
func (ch *CommandHandler) executeUse(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"使用什么？",
			"Shi yong shen me?",
			"Use what?",
		)
	}

	itemName := cmd.Args[0]

	var itemID string
	for _, id := range ch.gameState.Player.Inventory {
		if item, exists := ch.gameState.Items[id]; exists {
			if item.Name.Chinese == itemName || strings.EqualFold(item.Name.English, itemName) {
				itemID = id
				break
			}
		}
	}

	if itemID == "" {
		return "", ch.formatError(
			"你没有 "+itemName,
			"Ni mei you "+itemName,
			"You don't have "+itemName,
		)
	}

	// In executeMove, update error handling
	result, err := ch.gameState.SafeUseItem(itemID)
	if err != nil {
		if gameErr, ok := err.(*errors.GameError); ok {
			return "", fmt.Errorf(gameErr.GetUserMessage(
				ch.gameState.Config.ShouldShowChinese(),
				ch.gameState.Config.ShouldShowPinyin(),
				ch.gameState.Config.ShouldShowEnglish(),
			))
		}
		return "", err
	}

	return result, nil
}

// executeDrop handles dropping items
func (ch *CommandHandler) executeDrop(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"丢弃什么？",
			"Diu qi shen me?",
			"Drop what?",
		)
	}

	itemName := cmd.Args[0]

	var itemID string
	var foundItem models.Item
	for _, id := range ch.gameState.Player.Inventory {
		if item, exists := ch.gameState.Items[id]; exists {
			if item.Name.Chinese == itemName || strings.EqualFold(item.Name.English, itemName) {
				itemID = id
				foundItem = item
				break
			}
		}
	}

	if itemID == "" {
		return "", ch.formatError(
			"你没有 "+itemName,
			"Ni mei you "+itemName,
			"You don't have "+itemName,
		)
	}

	if ch.gameState.Player.GetEquippedWeaponID() == itemID {
		ch.gameState.Player.UnequipWeapon()
	}
	if ch.gameState.Player.GetEquippedArmorID() == itemID {
		ch.gameState.Player.UnequipArmor()
	}

	ch.gameState.Player.RemoveItem(itemID)
	ch.gameState.AddPendingDrop(ch.gameState.Player.CurrentLocationID, itemID)

	return ch.formatOutput(
		"你丢弃了 "+foundItem.Name.Chinese,
		"Ni diu qi le "+foundItem.Name.Pinyin,
		"You dropped: "+foundItem.Name.English,
	), nil
}

// executeEquipBackpack handles backpack equipment
func (ch *CommandHandler) executeEquipBackpack(cmd Command) (string, error) {
	if len(cmd.Args) == 0 {
		return "", ch.formatError(
			"带上哪个背包？",
			"Dai shang na ge bei bao?",
			"Equip which backpack?",
		)
	}

	itemName := cmd.Args[0]

	var itemID string
	for _, id := range ch.gameState.Player.Inventory {
		if item, exists := ch.gameState.Items[id]; exists {
			if item.IsBackpack() && (item.Name.Chinese == itemName || strings.EqualFold(item.Name.English, itemName)) {
				itemID = id
				break
			}
		}
	}

	if itemID == "" {
		return "", ch.formatError(
			"你没有这个背包",
			"Ni mei you zhe ge bei bao",
			"You don't have this backpack",
		)
	}

	return ch.gameState.SafeEquipBackpack(itemID)
}

// executeUnequipBackpack handles backpack unequipment
func (ch *CommandHandler) executeUnequipBackpack() (string, error) {
	result, err := ch.gameState.SafeUnequipBackpack()
	if err != nil {
		return "", err
	}
	// Replace the generic message with more natural Chinese
	return strings.Replace(result, "卸下了", "取下了", 1), nil
}
