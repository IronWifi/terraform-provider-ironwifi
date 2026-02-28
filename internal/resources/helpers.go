package resources

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// setIfNotNull sets a string value in the body map if the Terraform value is not null.
func setIfNotNull(body map[string]interface{}, key string, val types.String) {
	if !val.IsNull() && !val.IsUnknown() {
		body[key] = val.ValueString()
	}
}

// setIntIfNotNull sets an int64 value in the body map if the Terraform value is not null.
func setIntIfNotNull(body map[string]interface{}, key string, val types.Int64) {
	if !val.IsNull() && !val.IsUnknown() {
		body[key] = val.ValueInt64()
	}
}

// setBoolIfNotNull sets a bool value in the body map if the Terraform value is not null.
func setBoolIfNotNull(body map[string]interface{}, key string, val types.Bool) {
	if !val.IsNull() && !val.IsUnknown() {
		body[key] = val.ValueBool()
	}
}

// setBoolAsInt sets a bool as 0/1 integer in the body map (IronWiFi API convention).
func setBoolAsInt(body map[string]interface{}, key string, val types.Bool) {
	if !val.IsNull() && !val.IsUnknown() {
		if val.ValueBool() {
			body[key] = 1
		} else {
			body[key] = 0
		}
	}
}

// stringFromAPI extracts a string value from the API response map.
func stringFromAPI(data map[string]interface{}, key string) types.String {
	if v, ok := data[key]; ok && v != nil {
		return types.StringValue(fmt.Sprintf("%v", v))
	}
	return types.StringValue("")
}

// stringFromAPINullable extracts a string value, returning null if not present.
func stringFromAPINullable(data map[string]interface{}, key string) types.String {
	if v, ok := data[key]; ok && v != nil {
		s := fmt.Sprintf("%v", v)
		if s != "" {
			return types.StringValue(s)
		}
	}
	return types.StringNull()
}

// intFromAPI extracts an int64 value from the API response map.
func intFromAPI(data map[string]interface{}, key string) types.Int64 {
	if v, ok := data[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return types.Int64Value(int64(n))
		case int64:
			return types.Int64Value(n)
		case int:
			return types.Int64Value(int64(n))
		}
	}
	return types.Int64Value(0)
}

// intFromAPINullable extracts an int64 value, returning null if not present.
func intFromAPINullable(data map[string]interface{}, key string) types.Int64 {
	if v, ok := data[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return types.Int64Value(int64(n))
		case int64:
			return types.Int64Value(n)
		case int:
			return types.Int64Value(int64(n))
		}
	}
	return types.Int64Null()
}

// boolFromAPI extracts a bool value from the API response map.
func boolFromAPI(data map[string]interface{}, key string) types.Bool {
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

// boolFromIntAPI extracts a bool from a 0/1 integer API field.
func boolFromIntAPI(data map[string]interface{}, key string) types.Bool {
	if v, ok := data[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return types.BoolValue(n != 0)
		case int:
			return types.BoolValue(n != 0)
		case int64:
			return types.BoolValue(n != 0)
		case bool:
			return types.BoolValue(n)
		case string:
			return types.BoolValue(n == "1" || n == "true")
		}
	}
	return types.BoolValue(false)
}

// boolFromAPINullable extracts a bool, returning null if not present.
func boolFromAPINullable(data map[string]interface{}, key string) types.Bool {
	if v, ok := data[key]; ok && v != nil {
		switch n := v.(type) {
		case float64:
			return types.BoolValue(n != 0)
		case bool:
			return types.BoolValue(n)
		case string:
			return types.BoolValue(n == "1" || n == "true")
		}
	}
	return types.BoolNull()
}
