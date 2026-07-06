package main

import (
	"context"
	"fmt"
	"log"

	reference "github.com/cicbyte/reference/pkg"
)

func main() {
	// 创建 Engine 实例
	engine, err := reference.New(reference.Options{
		// 使用默认配置
	})
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	ctx := context.Background()

	// 列出当前项目的引用仓库
	result, err := engine.ListRepos(ctx, ".")
	if err != nil {
		log.Fatalf("Failed to list repos: %v", err)
	}

	fmt.Printf("Found %d repos:\n", len(result.Repos))
	for _, repo := range result.Repos {
		fmt.Printf("  - %s (%s): %s\n", repo.Name, repo.Type, repo.Source)
	}
}
