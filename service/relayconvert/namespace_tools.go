package relayconvert

import (
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/service/relayconvert/internal/shared/nstool"
)

// NamespacedTool identifies a function inside a Responses namespace tool.
type NamespacedTool = nstool.Tool

// NamespaceToolMapFromRequest returns flat chat tool name → namespaced tool
// for a client Responses request, or nil when the request is not a Responses
// request or declares no namespace tools.
func NamespaceToolMapFromRequest(request any) map[string]NamespacedTool {
	req, ok := request.(*dto.OpenAIResponsesRequest)
	if !ok || req == nil {
		return nil
	}
	return nstool.MapFromTools(req.Tools)
}

// RestoreNamespacedOutput rewrites a flattened function_call output item back
// to the namespaced (namespace, name) form expected by Responses API clients.
func RestoreNamespacedOutput(item *dto.ResponsesOutput, toolMap map[string]NamespacedTool) {
	nstool.RestoreOutput(item, toolMap)
}
