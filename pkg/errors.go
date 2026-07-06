package reference

import "errors"

var (
	// ErrEngineClosed Engine 已关闭
	ErrEngineClosed = errors.New("engine is closed")

	// ErrPromptsFSNotSet 嵌入的提示词文件系统未设置
	ErrPromptsFSNotSet = errors.New("prompts filesystem not set, call SetPromptsFS() first")

	// ErrInvalidOptions 无效的配置选项
	ErrInvalidOptions = errors.New("invalid options")
)
