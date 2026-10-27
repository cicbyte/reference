package repo

import (
	"strings"
	"testing"
)

const canonicalAgentTemplate = `---
name: reference-explorer
description: 专用于探索特定问题并积累知识
tools: Read, Grep, Glob, Bash, Write
---

# 角色：知识探索员

正文内容。
`

func TestTransformNativeToolsFrontmatter(t *testing.T) {
	got := string(transformNativeToolsFrontmatter([]byte(canonicalAgentTemplate), nativeToolNames))

	assert(t, true, strings.Contains(got, "name: reference-explorer\n"))
	assert(t, true, strings.Contains(got, "description: 专用于探索特定问题并积累知识\n"))
	assert(t, true, strings.Contains(got, "tools:\n  - read_file\n  - grep_search\n  - glob\n  - run_shell_command\n  - write_file\n"))
	assert(t, true, strings.Contains(got, "# 角色：知识探索员"))
	assert(t, false, strings.Contains(got, "tools: Read"))
}

func TestTransformNativeToolsFrontmatter_NoFrontmatter(t *testing.T) {
	data := []byte("# 只有正文\n没有 frontmatter\n")
	assert(t, string(data), string(transformNativeToolsFrontmatter(data, nativeToolNames)))
}

func TestTransformIflowFrontmatter(t *testing.T) {
	got := string(transformIflowFrontmatter([]byte(canonicalAgentTemplate)))

	assert(t, true, strings.Contains(got, `agentType: "reference-explorer"`))
	assert(t, true, strings.Contains(got, `whenToUse: "专用于探索特定问题并积累知识"`))
	assert(t, true, strings.Contains(got, "allowedTools:\n  - Read\n  - Grep\n  - Glob\n  - Bash\n  - Write\n"))
	assert(t, true, strings.Contains(got, "proactive: true"))
	assert(t, false, strings.Contains(got, "tools: Read"))
	// 正文完整保留
	assert(t, true, strings.Contains(got, "# 角色：知识探索员"))
}

func TestNewAgentsRegistered(t *testing.T) {
	for _, id := range []string{"gemini", "qwen", "kimi", "iflow"} {
		assert(t, true, IsAgentRegistered(id))
		cfg, ok := GetAgentConfig(id)
		assert(t, true, ok)
		assert(t, true, len(cfg.Files) > 0)
	}
}
