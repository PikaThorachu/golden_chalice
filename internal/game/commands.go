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

	if strings.HasPrefix(input, "往") {
		if dir, err := models.ParseDirection(input); err == nil {
			return Command{Type: CmdMove, Args: []string{dir.String()}, RawInput: input}
		}
	}

	if strings.HasPrefix(inputLower, "go ") || strings.HasPrefix(inputLower, "walk ") {
		if dir, err := models.ParseDirection(input); err == nil {
			return Command{Type: CmdMove, Args: []string{dir.String()}, RawInput: input}
		}
	}

	if dir, err := models.ParseDirection(input); err == nil {
		return Command{Type: CmdMove, Args: []string{dir.String()}, RawInput: input}
	}

	if strings.HasPrefix(input, "拿") || strings.HasPrefix(input, "取") {
		itemName := strings.TrimPrefix(input, "拿")
		itemName = strings.TrimPrefix(itemName, "取")
		itemName = strings.TrimSpace(itemName)
		if itemName != "" {
			return Command{Type: CmdTake, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	if strings.HasPrefix(inputLower, "take ") {
		itemName := strings.TrimPrefix(inputLower, "take ")
		itemName = strings.TrimSpace(itemName)
		if itemName != "" {
			return Command{Type: CmdTake, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	if input == "背包" || input == "i" || input == "inventory" {
		return Command{Type: CmdInventory, RawInput: input}
	}

	if input == "状态" || input == "status" {
		return Command{Type: CmdStatus, RawInput: input}
	}

	if input == "帮助" || input == "help" {
		return Command{Type: CmdHelp, RawInput: input}
	}

	if input == "退出" || input == "quit" || input == "exit" {
		return Command{Type: CmdQuit, RawInput: input}
	}

	if input == "看" || input == "查看" || input == "look" {
		return Command{Type: CmdLook, RawInput: input}
	}

	if strings.HasPrefix(inputLower, "save ") || strings.HasPrefix(input, "保存 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			slotName := parts[1]
			return Command{Type: CmdSave, Args: []string{slotName}, RawInput: input}
		}
		return Command{Type: CmdSave, Args: []string{"autosave"}, RawInput: input}
	}

	if strings.HasPrefix(inputLower, "load ") || strings.HasPrefix(input, "加载 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			slotName := parts[1]
			return Command{Type: CmdLoad, Args: []string{slotName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	if input == "saves" || input == "存档列表" {
		return Command{Type: CmdListSaves, RawInput: input}
	}

	if strings.HasPrefix(inputLower, "delete ") || strings.HasPrefix(input, "删除 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			slotName := parts[1]
			return Command{Type: CmdDeleteSave, Args: []string{slotName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	if strings.HasPrefix(inputLower, "equip ") || strings.HasPrefix(input, "装备 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			itemName := parts[1]
			if itemName == "backpack" || itemName == "背包" {
				return Command{Type: CmdEquipBackpack, Args: []string{itemName}, RawInput: input}
			}
			return Command{Type: CmdEquip, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	if strings.HasPrefix(inputLower, "unequip ") || strings.HasPrefix(input, "卸下 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			slot := parts[1]
			if slot == "backpack" || slot == "背包" {
				return Command{Type: CmdUnequipBackpack, RawInput: input}
			}
			return Command{Type: CmdUnequip, Args: []string{slot}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

	if strings.HasPrefix(inputLower, "use ") || strings.HasPrefix(input, "使用 ") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			itemName := parts[1]
			return Command{Type: CmdUse, Args: []string{itemName}, RawInput: input}
		}
		return Command{Type: CmdUnknown, RawInput: input}
	}

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
			return "", ch.formatError(
				gameErr.Message,
				gameErr.MessagePinyin,
				gameErr.MessageEnglish,
			)
		}
		return "", err
	}

	return ch.gameState.GetCurrentRoomDescription(), nil
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
			return "", ch.formatError(
				gameErr.Message,
				gameErr.MessagePinyin,
				gameErr.MessageEnglish,
			)
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

// executeStatus displays player status
func (ch *CommandHandler) executeStatus() (string, error) {
	return ch.gameState.GetPlayerStatus(), nil
}

// executeHelp displays help text
func (ch *CommandHandler) executeHelp() (string, error) {
	helpChinese := `╔══════════════════════════════════════════════════════════╗
║                      游戏命令帮助                         ║
╠══════════════════════════════════════════════════════════╣
║ 移动:                                                      ║
║   中文: 北, 往北, 往北走, 西北, 往西北走, 往西北去        ║
║   英文: go north, go northwest, walk south, etc.         ║
╠══════════════════════════════════════════════════════════╣
║ 物品操作:                                                  ║
║   拿取: 拿<物品名> 或 take <item>                         ║
║   使用: 使用<物品名> 或 use <item>                        ║
║   装备: 装备<物品名> 或 equip <item>                      ║
║   卸下: 卸下<武器/护甲> 或 unequip <weapon/armor>         ║
║   丢弃: 丢弃<物品名> 或 drop <item>                       ║
╠══════════════════════════════════════════════════════════╣
║ 信息查看:                                                  ║
║   背包: 背包 或 i 或 inventory                            ║
║   状态: 状态 或 status                                    ║
║   查看: 看 或 查看 或 look                                ║
║   帮助: 帮助 或 help                                      ║
╠══════════════════════════════════════════════════════════╣
║ 存档管理:                                                  ║
║   保存: save <名称> 或 保存 <名称>                        ║
║   加载: load <名称> 或 加载 <名称>                        ║
║   列表: saves 或 存档列表                                 ║
║   删除: delete <名称> 或 删除 <名称>                      ║
╠══════════════════════════════════════════════════════════╣
║ 其他:                                                      ║
║   退出: quit 或 退出                                      ║
╚══════════════════════════════════════════════════════════╝`

	helpPinyin := "=== You Xi Ming Ling Bang Zhu ==="
	helpEnglish := `╔══════════════════════════════════════════════════════════╗
║                      GAME COMMANDS                         ║
╠══════════════════════════════════════════════════════════╣
║ Movement:                                                   ║
║   Chinese: 北, 往北, 往北走, 西北, 往西北走               ║
║   English: go north, go northwest, walk south, etc.       ║
╠══════════════════════════════════════════════════════════╣
║ Items:                                                      ║
║   Take: 拿<item> or take <item>                           ║
║   Use: 使用<item> or use <item>                           ║
║   Equip: 装备<item> or equip <item>                       ║
║   Unequip: 卸下<weapon/armor> or unequip <weapon/armor>   ║
║   Drop: 丢弃<item> or drop <item>                         ║
╠══════════════════════════════════════════════════════════╣
║ Information:                                                ║
║   Inventory: 背包, i, or inventory                        ║
║   Status: 状态 or status                                  ║
║   Look: 看, 查看, or look                                 ║
║   Help: 帮助 or help                                      ║
╠══════════════════════════════════════════════════════════╣
║ Save/Load:                                                  ║
║   Save: save <name> or 保存 <name>                        ║
║   Load: load <name> or 加载 <name>                        ║
║   List: saves or 存档列表                                 ║
║   Delete: delete <name> or 删除 <name>                    ║
╠══════════════════════════════════════════════════════════╣
║ Other:                                                      ║
║   Quit: quit or 退出                                       ║
╚══════════════════════════════════════════════════════════╝`

	return ch.formatOutput(helpChinese, helpPinyin, helpEnglish), nil
}

// executeQuit handles quit command
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

	if foundItem.IsWeapon() {
		ch.gameState.Player.EquipWeapon(itemID)
		return ch.formatOutput(
			"装备了 "+foundItem.Name.Chinese,
			"Zhuang bei le "+foundItem.Name.Pinyin,
			"Equipped: "+foundItem.Name.English,
		), nil
	}

	if foundItem.IsArmor() {
		ch.gameState.Player.EquipArmor(itemID)
		return ch.formatOutput(
			"装备了 "+foundItem.Name.Chinese,
			"Zhuang bei le "+foundItem.Name.Pinyin,
			"Equipped: "+foundItem.Name.English,
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

	if slot == "武器" || slot == "weapon" {
		if ch.gameState.Player.HasEquippedWeapon() {
			ch.gameState.Player.UnequipWeapon()
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
			ch.gameState.Player.UnequipArmor()
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

	result, err := ch.gameState.SafeUseItem(itemID)
	if err != nil {
		if gameErr, ok := err.(*errors.GameError); ok {
			return "", ch.formatError(
				gameErr.Message,
				gameErr.MessagePinyin,
				gameErr.MessageEnglish,
			)
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
			"装备哪个背包？",
			"Zhuang bei na ge bei bao?",
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
	return ch.gameState.SafeUnequipBackpack()
}
