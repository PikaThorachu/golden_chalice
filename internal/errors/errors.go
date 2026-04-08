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

// GameError represents a structured error with context
type GameError struct {
	Type           ErrorType
	Message        string
	MessagePinyin  string
	MessageEnglish string
	Err            error
	Context        map[string]interface{}
}

// Error implements the error interface
func (e *GameError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *GameError) Unwrap() error {
	return e.Err
}

// New creates a new GameError
func New(errType ErrorType, message, pinyin, english string) *GameError {
	return &GameError{
		Type:           errType,
		Message:        message,
		MessagePinyin:  pinyin,
		MessageEnglish: english,
		Context:        make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, message, pinyin, english string) *GameError {
	return &GameError{
		Type:           errType,
		Message:        message,
		MessagePinyin:  pinyin,
		MessageEnglish: english,
		Err:            err,
		Context:        make(map[string]interface{}),
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

	if showChinese && e.Message != "" {
		parts = append(parts, e.Message)
	}
	if showPinyin && e.MessagePinyin != "" {
		pinyinText := e.MessagePinyin
		if showChinese && len(parts) > 0 {
			pinyinText = "(" + pinyinText + ")"
		}
		parts = append(parts, pinyinText)
	}
	if showEnglish && e.MessageEnglish != "" {
		englishText := e.MessageEnglish
		if len(parts) > 0 {
			englishText = "/ " + englishText
		}
		parts = append(parts, englishText)
	}

	if len(parts) == 0 {
		return e.Message
	}
	return strings.Join(parts, " ")
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
		"Shūrù bùnéng wéi kōng",
		"Input cannot be empty")

	ErrInvalidCommand = New(ErrTypeValidation,
		"无效的命令",
		"Wúxiào de mìnglìng",
		"Invalid command")

	// Movement errors
	ErrNoExit = New(ErrTypeMovement,
		"一堵石墙挡住了你的去路",
		"Yī dǔ shíqiáng dǎngzhùle nǐ de qùlù",
		"A stone wall blocks your path")

	ErrEnemyBlocks = New(ErrTypeMovement,
		"有敌人挡住了去路",
		"Yǒu dírén dǎngzhùle qùlù",
		"An enemy blocks your path")

	ErrExitLocked = New(ErrTypeMovement,
		"门被锁住了",
		"Mén bèi suǒ zhù le",
		"The door is locked")

	// Combat errors
	ErrAlreadyDefeated = New(ErrTypeCombat,
		"敌人已经被击败了",
		"Dírén yǐjīng bèi jībài le",
		"Enemy has already been defeated")

	ErrPlayerDead = New(ErrTypeCombat,
		"你已经死亡，无法战斗",
		"Nǐ yǐjīng sǐwáng, wúfǎ zhàndòu",
		"You are dead and cannot fight")

	// Inventory errors
	ErrItemNotFound = New(ErrTypeInventory,
		"物品不存在",
		"Wùpǐn bù cúnzài",
		"Item not found")

	ErrInventoryFull = New(ErrTypeInventory,
		"背包已满",
		"Bēibāo yǐ mǎn",
		"Inventory is full")

	ErrCannotEquip = New(ErrTypeInventory,
		"无法装备该物品",
		"Wúfǎ zhuāngbèi gāi wùpǐn",
		"Cannot equip this item")

	// Save/Load errors
	ErrSaveNotFound = New(ErrTypeSaveLoad,
		"存档不存在",
		"Cúndàng bù cúnzài",
		"Save file not found")

	ErrSaveCorrupted = New(ErrTypeSaveLoad,
		"存档已损坏",
		"Cúndàng yǐ sǔnhuài",
		"Save file is corrupted")

	ErrSaveFailed = New(ErrTypeSaveLoad,
		"保存失败",
		"Bǎocún shībài",
		"Failed to save")

	// Game state errors
	ErrGameAlreadyOver = New(ErrTypeGameState,
		"游戏已经结束",
		"Yóuxì yǐjīng jiéshù",
		"Game has already ended")

	ErrNoActiveGame = New(ErrTypeGameState,
		"没有进行中的游戏",
		"Méiyǒu jìnxíng zhōng de yóuxì",
		"No active game")

	// Configuration errors
	ErrConfigInvalid = New(ErrTypeConfiguration,
		"配置无效",
		"Pèizhì wúxiào",
		"Invalid configuration")
)
