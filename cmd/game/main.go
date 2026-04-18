package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"golden_chalice/internal/game"
	"golden_chalice/internal/loader"
	"golden_chalice/internal/logging"
	"golden_chalice/internal/models"
	"golden_chalice/internal/save"
)

// displayFormatter will be used for trilingual output
var displayFormatter *game.DisplayFormatter

func main() {
	// Initialize random seed for combat and drops
	rand.Seed(time.Now().UnixNano())

	// Display startup message
	printTrilingual(
		"正在启动游戏...",
		"Zheng zai qi dong you xi...",
		"Starting game...",
	)

	// Initialize logger
	logger, err := initializeLogger()
	if err != nil {
		printTrilingual(
			fmt.Sprintf("警告: 无法初始化日志系统: %v", err),
			fmt.Sprintf("Jing gao: Wu fa chu shi hua ri zhi xi tong: %v", err),
			fmt.Sprintf("Warning: Unable to initialize logging system: %v", err),
		)
		logger = nil
	} else {
		printTrilingual(
			"✓ 日志系统已初始化",
			"✓ Ri zhi xi tong yi chu shi hua",
			"✓ Logging system initialized",
		)
		defer logger.Close()
	}

	// Load all game data with validation
	gameData, err := loadGameData(logger)
	if err != nil {
		printTrilingual(
			fmt.Sprintf("❌ 游戏加载失败: %v", err),
			fmt.Sprintf("❌ You xi jia zai shi bai: %v", err),
			fmt.Sprintf("❌ Game load failed: %v", err),
		)
		printTrilingual(
			"请检查以下内容:",
			"Qing jian cha yi xia nei rong:",
			"Please check the following:",
		)
		printTrilingual(
			"  1. data/ 目录是否存在",
			"  1. data/ mu lu shi fou cun zai",
			"  1. data/ directory exists",
		)
		printTrilingual(
			"  2. 所有 JSON 文件是否格式正确",
			"  2. Suo you JSON wen jian shi fou ge shi zheng que",
			"  2. All JSON files are properly formatted",
		)
		printTrilingual(
			"  3. 配置文件中的引用是否有效",
			"  3. Pei zhi wen jian zhong de yin yong shi fou you xiao",
			"  3. References in config files are valid",
		)
		printTrilingual(
			"\n按 Enter 键退出...",
			"\nAn Enter jian tui chu...",
			"\nPress Enter to exit...",
		)
		bufio.NewReader(os.Stdin).ReadString('\n')
		os.Exit(1)
	}

	// Initialize display formatter
	displayFormatter = game.NewDisplayFormatter(gameData.Config)

	// Initialize save manager
	saveMgr, err := save.NewSaveManager("./saves", "./data/save_config.json")
	if err != nil {
		printTrilingual(
			fmt.Sprintf("警告: 保存系统初始化失败: %v", err),
			fmt.Sprintf("Jing gao: Bao cun xi tong chu shi hua shi bai: %v", err),
			fmt.Sprintf("Warning: Save system initialization failed: %v", err),
		)
		printTrilingual(
			"游戏将继续，但无法保存进度",
			"You xi jiang ji xu, dan wu fa bao cun jin du",
			"Game will continue, but cannot save progress",
		)
		saveMgr = nil
	} else {
		printTrilingual(
			"✓ 保存系统已初始化",
			"✓ Bao cun xi tong yi chu shi hua",
			"✓ Save system initialized",
		)
	}

	// Create game state
	gs := game.NewGameState(
		gameData.Config,
		gameData.World,
		gameData.Items,
		gameData.Enemies,
		gameData.Biomes,
		logger,
	)

	// Display welcome message
	fmt.Print(gs.Formatter.FormatWelcome(gameData.Config.GameVersion))

	// Main menu
	reader := bufio.NewReader(os.Stdin)
	for {
		printMainMenu()
		fmt.Print(getTrilingualInputPrompt())

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		// Check for quit commands from main menu
		if input == "退出" || input == "quit" || input == "exit" || input == "退" {
			printTrilingual(
				"\n感谢游玩！再见！",
				"\nGan xie you wan! Zai jian!",
				"\nThanks for playing! Farewell!",
			)
			return
		}

		switch input {
		case "1":
			startNewGame(gs, saveMgr, reader)
			return
		case "2":
			if saveMgr == nil {
				printTrilingual(
					"保存系统不可用，无法加载存档",
					"Bao cun xi tong bu ke yong, wu fa jia zai cun dang",
					"Save system unavailable, cannot load save",
				)
				continue
			}
			if loadGame(gs, saveMgr, reader) {
				runGameLoop(gs, saveMgr, reader)
				return
			}
		case "3":
			printTrilingual(
				"\n感谢游玩！再见！",
				"\nGan xie you wan! Zai jian!",
				"\nThanks for playing! Farewell!",
			)
			return
		default:
			printTrilingual(
				"无效选择，请重新输入",
				"Wu xiao xuan ze, qing chong xin shu ru",
				"Invalid choice, please try again",
			)
		}
	}
}

// initializeLogger creates and configures the game logger
func initializeLogger() (*logging.Logger, error) {
	config := logging.DefaultConfig()
	config.LogDirectory = "./logs"
	config.LogFileName = "game.log"
	config.Level = logging.LevelInfo
	config.ConsoleOutput = true
	config.FileOutput = true

	return logging.NewLogger(config)
}

// loadGameData loads and validates all game data
func loadGameData(logger *logging.Logger) (*loader.LoadedGameData, error) {
	printTrilingual(
		"\n正在加载游戏数据...",
		"\nZheng zai jia zai you xi shu ju...",
		"\nLoading game data...",
	)

	// Validate data directory exists
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		return nil, fmt.Errorf("data directory does not exist")
	}

	// Load all game data
	gameData, err := loader.LoadAllGameData("./data")
	if err != nil {
		if logger != nil {
			logger.LogError(err, map[string]interface{}{
				"phase": "load_game_data",
			})
		}
		return nil, err
	}

	// Additional validation
	if err := validateGameData(gameData); err != nil {
		return nil, err
	}

	return gameData, nil
}

// validateGameData performs additional cross-validation
func validateGameData(gameData *loader.LoadedGameData) error {
	// Validate starting location
	if _, err := gameData.World.GetLocation(gameData.Config.GetStartingLocationID()); err != nil {
		return fmt.Errorf("starting location '%s' not found in world data", gameData.Config.GetStartingLocationID())
	}

	// Validate win condition item
	winItemID := gameData.Config.GetWinConditionItemID()
	if _, err := loader.GetItem(gameData.Items, winItemID); err != nil {
		return fmt.Errorf("win condition item '%s' not found in items data", winItemID)
	}

	// Validate starting health is reasonable
	if gameData.Config.GetStartingHealth() <= 0 {
		return fmt.Errorf("invalid starting health: %d", gameData.Config.GetStartingHealth())
	}

	// Validate inventory settings
	if gameData.Config.GetStartingInventorySize() <= 0 {
		return fmt.Errorf("invalid starting inventory size: %d", gameData.Config.GetStartingInventorySize())
	}

	if gameData.Config.GetMaxInventorySize() < gameData.Config.GetStartingInventorySize() {
		return fmt.Errorf("max inventory size (%d) is less than starting size (%d)",
			gameData.Config.GetMaxInventorySize(),
			gameData.Config.GetStartingInventorySize())
	}

	// Validate combat settings
	if gameData.Config.CombatSettings.BaseDamageMin <= 0 {
		return fmt.Errorf("invalid base damage min: %d", gameData.Config.CombatSettings.BaseDamageMin)
	}
	if gameData.Config.CombatSettings.BaseDamageMax <= 0 {
		return fmt.Errorf("invalid base damage max: %d", gameData.Config.CombatSettings.BaseDamageMax)
	}
	if gameData.Config.CombatSettings.BaseDamageMin > gameData.Config.CombatSettings.BaseDamageMax {
		return fmt.Errorf("base damage min (%d) is greater than max (%d)",
			gameData.Config.CombatSettings.BaseDamageMin,
			gameData.Config.CombatSettings.BaseDamageMax)
	}

	printTrilingual(
		"✓ 所有游戏数据验证通过",
		"✓ Suo you you xi shu ju yan zheng tong guo",
		"✓ All game data validated successfully",
	)
	return nil
}

// startNewGame begins a new game
func startNewGame(gs *game.GameState, saveMgr *save.SaveManager, reader *bufio.Reader) {
	printTrilingual(
		"\n请输入你的名字: ",
		"\nQing shu ru ni de ming zi: ",
		"\nPlease enter your name: ",
	)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		name = "冒险者"
		printTrilingual(
			"使用默认名字: 冒险者",
			"Shi yong mo ren ming zi: Mao Xian Zhe",
			"Using default name: Adventurer",
		)
	}

	gs.NewGame(name)
	printTrilingual(
		fmt.Sprintf("\n欢迎, %s!", name),
		fmt.Sprintf("\nHuan ying, %s!", name),
		fmt.Sprintf("\nWelcome, %s!", name),
	)
	printTrilingual(
		"游戏开始...",
		"You xi kai shi...",
		"Game starting...",
	)
	printTrilingual(
		"提示: 输入 '帮助' 查看可用命令",
		"Ti shi: Shu ru 'bang zhu' cha kan ke yong ming ling",
		"Hint: Type 'help' to see available commands",
	)

	runGameLoop(gs, saveMgr, reader)
}

// loadGame handles loading a saved game
func loadGame(gs *game.GameState, saveMgr *save.SaveManager, reader *bufio.Reader) bool {
	// List available saves
	saves, err := saveMgr.ListSaves()
	if err != nil {
		printTrilingual(
			fmt.Sprintf("无法列出存档: %v", err),
			fmt.Sprintf("Wu fa lie chu cun dang: %v", err),
			fmt.Sprintf("Failed to list saves: %v", err),
		)
		return false
	}

	if len(saves) == 0 {
		printTrilingual(
			"没有找到任何存档",
			"Mei you zhao dao ren he cun dang",
			"No saves found",
		)
		return false
	}

	printTrilingual(
		"\n可用存档:",
		"\nKe yong cun dang:",
		"\nAvailable saves:",
	)
	for i, s := range saves {
		slotDisplay := fmt.Sprintf("  %d. %s - %s (%s)", i+1, s.DisplayName, s.PlayerName, s.SaveTime.Format("2006-01-02 15:04:05"))
		printTrilingual(slotDisplay, slotDisplay, slotDisplay)
	}
	printTrilingual(
		"  0. 返回主菜单",
		"  0. Fan hui zhu cai dan",
		"  0. Return to main menu",
	)

	printTrilingual(
		"\n请选择存档: ",
		"\nQing xuan ze cun dang: ",
		"\nSelect save: ",
	)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "0" {
		return false
	}

	var selectedSave save.SaveInfo
	switch input {
	case "1":
		if len(saves) > 0 {
			selectedSave = saves[0]
		}
	case "2":
		if len(saves) > 1 {
			selectedSave = saves[1]
		}
	case "3":
		if len(saves) > 2 {
			selectedSave = saves[2]
		}
	default:
		printTrilingual(
			"无效选择",
			"Wu xiao xuan ze",
			"Invalid choice",
		)
		return false
	}

	if selectedSave.SlotName == "" {
		printTrilingual(
			"无效选择",
			"Wu xiao xuan ze",
			"Invalid choice",
		)
		return false
	}

	printTrilingual(
		fmt.Sprintf("正在加载存档 %s...", selectedSave.DisplayName),
		fmt.Sprintf("Zheng zai jia zai cun dang %s...", selectedSave.DisplayName),
		fmt.Sprintf("Loading save %s...", selectedSave.DisplayName),
	)

	saveData, err := saveMgr.LoadFromFile(selectedSave.SlotName)
	if err != nil {
		printTrilingual(
			fmt.Sprintf("加载失败: %v", err),
			fmt.Sprintf("Jia zai shi bai: %v", err),
			fmt.Sprintf("Load failed: %v", err),
		)
		return false
	}

	// Restore game state
	gs.Player = saveData.Player
	gs.DefeatedEnemies = saveData.DefeatedEnemies
	gs.TakenItems = saveData.TakenItems
	gs.PendingDrops = saveData.PendingDrops
	gs.GameOver = false
	gs.GameWon = false

	// Reset room tracking
	gs.CurrentRoomID = nil
	gs.LastLocationID = ""

	printTrilingual(
		fmt.Sprintf("欢迎回来, %s!", gs.Player.Name),
		fmt.Sprintf("Huan ying hui lai, %s!", gs.Player.Name),
		fmt.Sprintf("Welcome back, %s!", gs.Player.Name),
	)
	return true
}

// runGameLoop is the main game loop
func runGameLoop(gs *game.GameState, saveMgr *save.SaveManager, reader *bufio.Reader) {
	cmdHandler := game.NewCommandHandler(gs, saveMgr)
	lastSaveTime := time.Now()
	autoSaveInterval := time.Duration(gs.Config.GetAutoSaveInterval()) * time.Minute

	for !gs.GameOver {
		// Show current room
		fmt.Print(gs.GetCurrentRoomDescription())

		// Handle combat if enemy present
		if gs.HasEnemyAtCurrentLocation() {
			enemy, _ := gs.GetCurrentEnemy()
			fmt.Printf("\n⚔️ %s %s ⚔️\n",
				gs.GetDisplayName(enemy.Name),
				getTrilingualString("出现了！", "Chu xian le!", "appears!"))

			// Combat loop
			combatResult := runCombat(gs, enemy, reader)
			if combatResult == "defeated" {
				fmt.Println(gs.Formatter.FormatGameOver())
				break
			}
			if combatResult == "victory" {
				continue
			}
		}

		// Check win condition
		if gs.CheckWinCondition() {
			fmt.Print(gs.GetCurrentRoomDescription())
			fmt.Println(gs.Formatter.FormatVictory())
			printTrilingual(
				"你成功取得了金杯！",
				"Ni cheng gong qu de le jin bei!",
				"You have successfully obtained the Golden Chalice!",
			)
			break
		}

		// Check loss condition
		if gs.CheckLossCondition() {
			fmt.Println(gs.Formatter.FormatGameOver())
			break
		}

		// Auto-save
		if saveMgr != nil && gs.Config.IsAutoSaveEnabled() && time.Since(lastSaveTime) >= autoSaveInterval {
			err := saveMgr.AutoSave(
				gs.Player,
				gs.DefeatedEnemies,
				gs.TakenItems,
				gs.PendingDrops,
				gs.Config.GameVersion,
				gs.Player.Name,
			)
			if err == nil {
				printTrilingual(
					"\n[自动保存] 游戏已自动保存",
					"\n[Zi dong bao cun] You xi yi zi dong bao cun",
					"\n[Auto-save] Game has been auto-saved",
				)
				lastSaveTime = time.Now()
			}
		}

		// Get player input
		fmt.Print("\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Check for quit commands first
		lowerInput := strings.ToLower(input)
		if input == "退出" || input == "quit" || input == "exit" || input == "退" || lowerInput == "quit" || lowerInput == "exit" {
			// Confirm quit
			printTrilingual(
				"确定要退出吗？(y/n): ",
				"Que ding yao tui chu ma? (y/n): ",
				"Are you sure you want to quit? (y/n): ",
			)
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(strings.ToLower(confirm))
			if confirm == "y" || confirm == "yes" || confirm == "是" {
				// Auto-save before quitting
				if saveMgr != nil && gs.Config.IsAutoSaveEnabled() {
					err := saveMgr.AutoSave(
						gs.Player,
						gs.DefeatedEnemies,
						gs.TakenItems,
						gs.PendingDrops,
						gs.Config.GameVersion,
						gs.Player.Name,
					)
					if err == nil {
						printTrilingual(
							"游戏已自动保存。",
							"You xi yi zi dong bao cun.",
							"Game has been auto-saved.",
						)
					}
				}
				printTrilingual(
					"\n感谢游玩！再见！",
					"\nGan xie you wan! Zai jian!",
					"\nThanks for playing! Farewell!",
				)
				gs.GameOver = true
				return
			} else {
				printTrilingual(
					"继续游戏。",
					"Ji xu you xi.",
					"Continuing game.",
				)
				continue
			}
		}

		// Process command
		result, err := cmdHandler.ProcessCommand(input)
		if err != nil {
			fmt.Println(err)
		} else if result != "" {
			fmt.Println(result)
		}
	}

	printTrilingual(
		"\n游戏结束！感谢游玩！",
		"\nYou xi jie shu! Gan xie you wan!",
		"\nGame over! Thanks for playing!",
	)
}

// runCombat handles the combat sequence
func runCombat(gs *game.GameState, enemy *models.Enemy, reader *bufio.Reader) string {
	for {
		printTrilingual(
			"\n(攻)击 或 逃(跑)？ ",
			"\n(Gong)ji huo (Pao)pao? ",
			"\n(A)ttack or (R)un? ",
		)
		action, _ := reader.ReadString('\n')
		action = strings.TrimSpace(strings.ToLower(action))

		switch action {
		case "攻", "攻击", "a":
			victory, defeat, msg := gs.ProcessCombatTurn(enemy, true)
			fmt.Println(msg)

			if victory {
				printTrilingual(
					"战斗胜利！",
					"Zhan dou sheng li!",
					"Victory!",
				)
				return "victory"
			}
			if defeat {
				return "defeated"
			}

			victory, defeat, msg = gs.ProcessCombatTurn(enemy, false)
			fmt.Println(msg)

			if defeat {
				return "defeated"
			}

		case "跑", "逃跑", "r":
			success, msg := gs.AttemptFlee()
			fmt.Println(msg)
			if success {
				printTrilingual(
					"你成功逃跑了！",
					"Ni cheng gong tao pao le!",
					"You successfully fled!",
				)
				return "fled"
			}

		default:
			printTrilingual(
				"无效命令。请输入 '攻' 或 '跑'",
				"Wu xiao ming ling. Qing shu ru 'gong' huo 'pao'",
				"Invalid command. Please enter 'A' or 'R'",
			)
		}
	}
}

// cmd/game/main.go - Updated printMainMenu

func printMainMenu() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                       主菜单                              ║")

	// Add Pinyin if enabled in config
	if displayFormatter != nil && displayFormatter.ShowPinyin() {
		fmt.Println("║                       Zhu Cai Dan                         ║")
	}

	// Add English if enabled in config
	if displayFormatter != nil && displayFormatter.ShowEnglish() {
		fmt.Println("║                      Main Menu                            ║")
	}

	fmt.Println("╠═══════════════════════════════════════════════════════════╣")
	fmt.Println("║  1. 新游戏                                                ║")

	if displayFormatter != nil && displayFormatter.ShowPinyin() {
		fmt.Println("║      Xin You Xi                                           ║")
	}

	if displayFormatter != nil && displayFormatter.ShowEnglish() {
		fmt.Println("║      New Game                                             ║")
	}

	fmt.Println("║  2. 加载存档                                              ║")

	if displayFormatter != nil && displayFormatter.ShowPinyin() {
		fmt.Println("║      Jia Zai Cun Dang                                     ║")
	}

	if displayFormatter != nil && displayFormatter.ShowEnglish() {
		fmt.Println("║      Load Game                                            ║")
	}

	fmt.Println("║  3. 退出                                                  ║")

	if displayFormatter != nil && displayFormatter.ShowPinyin() {
		fmt.Println("║      Tui Chu                                              ║")
	}

	if displayFormatter != nil && displayFormatter.ShowEnglish() {
		fmt.Println("║      Quit                                                 ║")
	}

	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
}

// printTrilingual prints a message in Chinese, Pinyin, and English
func printTrilingual(chinese, pinyin, english string) {
	if displayFormatter != nil {
		text := models.Text{
			Chinese: chinese,
			Pinyin:  pinyin,
			English: english,
		}
		fmt.Println(displayFormatter.FormatText(text))
	} else {
		fmt.Println(chinese)
	}
}

// getTrilingualString returns a formatted trilingual string
func getTrilingualString(chinese, pinyin, english string) string {
	if displayFormatter != nil {
		text := models.Text{
			Chinese: chinese,
			Pinyin:  pinyin,
			English: english,
		}
		return displayFormatter.FormatText(text)
	}
	return chinese
}

// getTrilingualInputPrompt returns a formatted input prompt
func getTrilingualInputPrompt() string {
	if displayFormatter != nil {
		text := models.Text{
			Chinese: "请选择 (1-3): ",
			Pinyin:  "Qing xuan ze (1-3): ",
			English: "Choose (1-3): ",
		}
		return displayFormatter.FormatText(text)
	}
	return "请选择 (1-3): "
}
