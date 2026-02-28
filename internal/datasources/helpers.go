package datasources

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stringVal extracts a string value from an API response map.
func stringVal(data map[string]interface{}, key string) types.String {
	if v, ok := data[key]; ok && v != nil {
		return types.StringValue(fmt.Sprintf("%v", v))
	}
	return types.StringValue("")
}

// intVal extracts an int64 value from an API response map.
func intVal(data map[string]interface{}, key string) types.Int64 {
	if v, ok := data[key]; ok && v != nil {
		if n, ok := v.(float64); ok {
			return types.Int64Value(int64(n))
		}
	}
	return types.Int64Value(0)
}

// boolVal extracts a bool value from an API response map.
func boolVal(data map[string]interface{}, key string) types.Bool {
	if v, ok := data[key]; ok && v != nil {
		switch b := v.(type) {
		case bool:
			return types.BoolValue(b)
		case float64:
			return types.BoolValue(b != 0)
		case string:
			return types.BoolValue(b == "true" || b == "1")
		}
	}
	return types.BoolValue(false)
}
