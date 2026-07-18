// Package nstool flattens Responses API namespace tools (Codex multi-agent
// groups, MCP servers) into plain Chat Completions function tools and restores
// the (namespace, name) split on the way back. Codex routes tool calls by an
// exact (namespace, name) match, so both directions must agree on the naming.
package nstool

import (
	"encoding/json"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

const separator = "__"

// Tool identifies a function inside a Responses namespace tool.
type Tool struct {
	Namespace string
	Name      string
}

// FlatName builds the Chat Completions tool name for a namespaced function.
func FlatName(namespace, name string) string {
	return namespace + separator + name
}

// IsNamespaceType reports whether a Responses tool type groups inner function tools.
func IsNamespaceType(toolType string) bool {
	return toolType == "namespace" || toolType == "mcp_server"
}

// MapFromTools parses a Responses API tools array and returns flat chat tool
// name → namespaced tool for every function inside namespace-like tools.
// Returns nil when the request declares no namespace tools.
func MapFromTools(raw json.RawMessage) map[string]Tool {
	if len(raw) == 0 {
		return nil
	}
	var tools []map[string]any
	if err := common.Unmarshal(raw, &tools); err != nil {
		return nil
	}
	out := map[string]Tool{}
	for _, tool := range tools {
		if !IsNamespaceType(strings.TrimSpace(common.Interface2String(tool["type"]))) {
			continue
		}
		namespace := strings.TrimSpace(common.Interface2String(tool["name"]))
		inner, ok := tool["tools"].([]any)
		if namespace == "" || !ok {
			continue
		}
		for _, item := range inner {
			innerTool, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name := strings.TrimSpace(common.Interface2String(innerTool["name"]))
			if name == "" {
				continue
			}
			out[FlatName(namespace, name)] = Tool{Namespace: namespace, Name: name}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// RestoreOutput rewrites a flattened function_call output item back to its
// namespaced form when its name matches a known namespace tool.
func RestoreOutput(item *dto.ResponsesOutput, toolMap map[string]Tool) {
	if item == nil || item.Type != "function_call" || len(toolMap) == 0 {
		return
	}
	if tool, ok := toolMap[item.Name]; ok {
		item.Name = tool.Name
		item.Namespace = tool.Namespace
	}
}
