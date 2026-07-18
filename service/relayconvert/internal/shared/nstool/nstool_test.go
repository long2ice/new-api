package nstool

import (
	"strings"

	"testing"

	"github.com/QuantumNous/new-api/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapFromToolsAndRestoreOutput(t *testing.T) {
	toolMap := MapFromTools([]byte(`[
		{"type": "function", "name": "lookup"},
		{
			"type": "namespace",
			"name": "multi_agent_v1",
			"tools": [
				{"type": "function", "name": "spawn_agent"},
				{"type": "function", "name": "wait_agent"}
			]
		},
		{"type": "web_search"}
	]`))
	require.Len(t, toolMap, 2)
	assert.Equal(t, Tool{Namespace: "multi_agent_v1", Name: "spawn_agent"}, toolMap["multi_agent_v1__spawn_agent"])

	item := &dto.ResponsesOutput{Type: "function_call", Name: "multi_agent_v1__spawn_agent"}
	RestoreOutput(item, toolMap)
	assert.Equal(t, "spawn_agent", item.Name)
	assert.Equal(t, "multi_agent_v1", item.Namespace)

	plain := &dto.ResponsesOutput{Type: "function_call", Name: "lookup"}
	RestoreOutput(plain, toolMap)
	assert.Equal(t, "lookup", plain.Name)
	assert.Empty(t, plain.Namespace)
}

func TestFlatNameCapsLongNames(t *testing.T) {
	short := FlatName("multi_agent_v1", "spawn_agent")
	assert.Equal(t, "multi_agent_v1__spawn_agent", short)

	namespace := strings.Repeat("a", 50)
	name := strings.Repeat("b", 50)
	long := FlatName(namespace, name)
	assert.Len(t, long, 64)
	// Deterministic: same inputs always produce the same capped name.
	assert.Equal(t, long, FlatName(namespace, name))
	assert.NotEqual(t, long, FlatName(namespace, strings.Repeat("c", 50)))
}

func TestMapFromToolsNoNamespaceTools(t *testing.T) {
	assert.Nil(t, MapFromTools(nil))
	assert.Nil(t, MapFromTools([]byte(`[{"type": "function", "name": "lookup"}]`)))
}
