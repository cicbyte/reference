package main

import (
	"embed"

	"github.com/cicbyte/reference/cmd"
	"github.com/cicbyte/reference/internal/common"
)

//go:embed prompts/**
var PromptsFS embed.FS

func main() {
	common.PromptsFS = PromptsFS
	cmd.Execute()
}
