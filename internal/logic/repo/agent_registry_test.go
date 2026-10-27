package repo

import (
	"sort"
	"testing"
)

func TestListAgentIDs_StableSortedOrder(t *testing.T) {
	ids := ListAgentIDs()

	assert(t, len(AgentRegistry), len(ids))
	assert(t, true, sort.StringsAreSorted(ids))
	for _, id := range ids {
		assert(t, true, IsAgentRegistered(id))
	}

	// 多次调用顺序必须一致，保证菜单编号稳定
	again := ListAgentIDs()
	for i := range ids {
		assert(t, ids[i], again[i])
	}
}

func TestValidateAgentIDs(t *testing.T) {
	valid := ValidateAgentIDs([]string{"claude", "zcode", " "})
	assert(t, 0, len(valid))

	invalid := ValidateAgentIDs([]string{"claude", "unknown-agent"})
	assert(t, 1, len(invalid))
}
