package sqidcty

import "github.com/zclconf/go-cty/cty/function"

// GetSqidFunctions returns all sqid cty functions for registration
// in an HCL2 eval context.
func GetSqidFunctions() map[string]function.Function {
	return map[string]function.Function{
		"sqid":   SqidFunc,
		"unsqid": UnsqidFunc,
	}
}
