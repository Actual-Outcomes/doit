package api

import (
	"context"
	"fmt"
	"time"

	"github.com/Actual-Outcomes/doit/internal/compact"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type compactArgs struct {
	Age string `json:"age"`
}

func (h *Handlers) Compact(ctx context.Context, _ *mcp.CallToolRequest, args compactArgs) (*mcp.CallToolResult, any, error) {
	age := args.Age
	if age == "" {
		age = "168h" // 7 days default
	}

	threshold, err := time.ParseDuration(age)
	if err != nil {
		return errResult(fmt.Errorf("invalid age duration %q: %w", age, err))
	}

	compactor := compact.New(h.store)
	results, err := compactor.CompactOld(ctx, threshold)
	if err != nil {
		return errResult(err)
	}

	return jsonResult(results)
}
