package errors

import (
	"fmt"
	"strings"
)

// ErrorType categorizes different kinds of errors
type ErrorType int

const (
	ErrTypeUnknown ErrorType = iota
	ErrTypeValidation
	ErrTypeNotFound
	ErrTypePermission
	ErrTypeCombat
	ErrTypeMovement
	ErrTypeInventory
	ErrTypeSaveLoad
	ErrTypeGameState
	ErrTypeConfiguration
)

// ErrorMessage holds trilingual error message text
type ErrorMessage struct {
	Chinese string
	Pinyin  string
	English string
}

// GameError represents a structured error with context
type GameError struct {
	Type    ErrorType
	Message ErrorMessage
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *GameError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message.Chinese, e.Err)
	}
	return e.Message.Chinese
}

// Unwrap returns the underlying error
func (e *GameError) Unwrap() error {
	return e.Err
}

// New creates a new GameError
func New(errType ErrorType, chinese, pinyin, english string) *GameError {
	return &GameError{
		Type: errType,
		Message: ErrorMessage{
			Chinese: chinese,
			Pinyin:  pinyin,
			English: english,
		},
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, chinese, pinyin, english string) *GameError {
	return &GameError{
		Type: errType,
		Message: ErrorMessage{
			Chinese: chinese,
			Pinyin:  pinyin,
			English: english,
		},
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *GameError) WithContext(key string, value interface{}) *GameError {
	e.Context[key] = value
	return e
}

// GetUserMessage returns the appropriate message based on language preferences
func (e *GameError) GetUserMessage(showChinese, showPinyin, showEnglish bool) string {
	var parts []string

	if showChinese && e.Message.Chinese != "" {
		parts = append(parts, e.Message.Chinese)
	}

	if showPinyin && e.Message.Pinyin != "" {
		pinyinText := e.Message.Pinyin
		if showChinese && len(parts) > 0 {
			pinyinText = "(" + pinyinText + ")"
		}
		parts = append(parts, pinyinText)
	}

	if showEnglish && e.Message.English != "" {
		englishText := e.Message.English
		if len(parts) > 0 {
			englishText = "/ " + englishText
		}
		parts = append(parts, englishText)
	}

	if len(parts) == 0 {
		return e.Message.Chinese
	}

	return strings.Join(parts, " ")
}

// ToError returns a standard error with the formatted message
func (e *GameError) ToError(showChinese, showPinyin, showEnglish bool) error {
	return fmt.Errorf("%s", e.GetUserMessage(showChinese, showPinyin, showEnglish))
}

// IsRecoverable checks if the error is recoverable (player can continue)
func (e *GameError) IsRecoverable() bool {
	switch e.Type {
	case ErrTypeValidation, ErrTypeMovement, ErrTypeInventory, ErrTypeCombat:
		return true
	case ErrTypeSaveLoad, ErrTypeGameState, ErrTypeConfiguration, ErrTypeNotFound, ErrTypePermission:
		return false
	default:
		return false
	}
}

// Predefined errors for common situations
var (
	// Validation errors
	ErrEmptyInput = New(ErrTypeValidation,
		"输入不能为空",
		"Shu ru bu neng wei kong",
		"Input cannot be empty")

	ErrInvalidCommand = New(ErrTypeValidation,
		"无效的命令",
		"Wu xiao de ming ling",
		"Invalid command")

	// Movement errors
	ErrNoExit = New(ErrTypeMovement,
		"一堵石墙挡住了你的去路",
		"Yi du shi qiang dang zhu le ni de qu lu",
		"A stone wall blocks your path")

	ErrEnemyBlocks = New(ErrTypeMovement,
		"有敌人挡住了去路",
		"You di ren dang zhu le qu lu",
		"An enemy blocks your path")

	ErrExitLocked = New(ErrTypeMovement,
		"门被锁住了",
		"Men bei suo zhu le",
		"The door is locked")

	// Combat errors
	ErrAlreadyDefeated = New(ErrTypeCombat,
		"敌人已经被击败了",
		"Di ren yi jing bei ji bai le",
		"Enemy has already been defeated")

	ErrPlayerDead = New(ErrTypeCombat,
		"你已经死亡，无法战斗",
		"Ni yi jing si wang, wu fa zhan dou",
		"You are dead and cannot fight")

	// Inventory errors
	ErrItemNotFound = New(ErrTypeInventory,
		"物品不存在",
		"Wu pin bu cun zai",
		"Item not found")

	ErrInventoryFull = New(ErrTypeInventory,
		"背包已满",
		"Bei bao yi man",
		"Inventory is full")

	ErrCannotEquip = New(ErrTypeInventory,
		"无法装备该物品",
		"Wu fa zhuang bei gai wu pin",
		"Cannot equip this item")

	// Save/Load errors
	ErrSaveNotFound = New(ErrTypeSaveLoad,
		"存档不存在",
		"Cun dang bu cun zai",
		"Save file not found")

	ErrSaveCorrupted = New(ErrTypeSaveLoad,
		"存档已损坏",
		"Cun dang yi sun huai",
		"Save file is corrupted")

	ErrSaveFailed = New(ErrTypeSaveLoad,
		"保存失败",
		"Bao cun shi bai",
		"Failed to save")

	// Game state errors
	ErrGameAlreadyOver = New(ErrTypeGameState,
		"游戏已经结束",
		"You xi yi jing jie shu",
		"Game has already ended")

	ErrNoActiveGame = New(ErrTypeGameState,
		"没有进行中的游戏",
		"Mei you jin xing zhong de you xi",
		"No active game")

	// Configuration errors
	ErrConfigInvalid = New(ErrTypeConfiguration,
		"配置无效",
		"Pei zhi wu xiao",
		"Invalid configuration")
)
