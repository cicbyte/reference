package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatJSONL OutputFormat = "jsonl"
)

func ParseFormat(s string) OutputFormat {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "jsonl":
		return FormatJSONL
	default:
		return FormatTable
	}
}

func OutputJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON 编码失败: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func OutputJSONL(items any) {
	data, err := json.Marshal(items)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON 编码失败: %v\n", err)
		return
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		fmt.Fprintf(os.Stderr, "JSON 解析失败: %v\n", err)
		return
	}
	for _, item := range arr {
		fmt.Println(string(item))
	}
}
