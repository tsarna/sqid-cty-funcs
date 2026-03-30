package sqidcty

import (
	"fmt"
	"math/big"

	sqids "github.com/sqids/sqids-go"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// SqidFunc encodes one or more non-negative integers into a sqid string.
// Called as sqid(id) or sqid(id, options).
// id may be a single number or a list of numbers.
var SqidFunc = function.New(&function.Spec{
	Description: "Encodes one or more non-negative integers into a sqid string",
	Params: []function.Parameter{
		{
			Name:             "id",
			Type:             cty.DynamicPseudoType,
			AllowDynamicType: true,
		},
	},
	VarParam: &function.Parameter{
		Name:             "options",
		Type:             cty.DynamicPseudoType,
		AllowDynamicType: true,
		AllowNull:        true,
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		if len(args) > 2 {
			return cty.NilType, fmt.Errorf("sqid() takes 1 or 2 arguments")
		}
		idType := args[0].Type()
		if idType != cty.DynamicPseudoType {
			if idType != cty.Number {
				if !idType.IsListType() || idType.ElementType() != cty.Number {
					return cty.NilType, fmt.Errorf(
						"sqid: id must be a number or list of numbers, got %s",
						idType.FriendlyName(),
					)
				}
			}
		}
		return cty.String, nil
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		var numbers []uint64

		switch args[0].Type() {
		case cty.Number:
			n, err := ctyNumberToUint64(args[0], "sqid")
			if err != nil {
				return cty.NilVal, err
			}
			numbers = []uint64{n}
		default:
			if !args[0].Type().IsListType() {
				return cty.NilVal, fmt.Errorf("sqid: id must be a number or list of numbers")
			}
			for it := args[0].ElementIterator(); it.Next(); {
				_, v := it.Element()
				n, err := ctyNumberToUint64(v, "sqid")
				if err != nil {
					return cty.NilVal, err
				}
				numbers = append(numbers, n)
			}
		}

		var optVal cty.Value
		if len(args) > 1 {
			optVal = args[1]
		}
		opts, err := parseSqidOptions(optVal)
		if err != nil {
			return cty.NilVal, err
		}

		s, err := sqids.New(opts)
		if err != nil {
			return cty.NilVal, fmt.Errorf("sqid: invalid options: %s", err)
		}

		result, err := s.Encode(numbers)
		if err != nil {
			return cty.NilVal, fmt.Errorf("sqid: encode error: %s", err)
		}

		return cty.StringVal(result), nil
	},
})

// UnsqidFunc decodes a sqid string into a list of non-negative integers.
// Called as unsqid(s) or unsqid(s, options).
// Returns an empty list for invalid or unrecognized input (by design).
var UnsqidFunc = function.New(&function.Spec{
	Description: "Decodes a sqid string into a list of non-negative integers; returns an empty list for invalid input",
	Params: []function.Parameter{
		{Name: "s", Type: cty.String},
	},
	VarParam: &function.Parameter{
		Name:             "options",
		Type:             cty.DynamicPseudoType,
		AllowDynamicType: true,
		AllowNull:        true,
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		if len(args) > 2 {
			return cty.NilType, fmt.Errorf("unsqid() takes 1 or 2 arguments")
		}
		return cty.List(cty.Number), nil
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		id := args[0].AsString()

		var optVal cty.Value
		if len(args) > 1 {
			optVal = args[1]
		}
		opts, err := parseSqidOptions(optVal)
		if err != nil {
			return cty.NilVal, err
		}

		s, err := sqids.New(opts)
		if err != nil {
			return cty.NilVal, fmt.Errorf("unsqid: invalid options: %s", err)
		}

		numbers := s.Decode(id)

		if len(numbers) == 0 {
			return cty.ListValEmpty(cty.Number), nil
		}

		vals := make([]cty.Value, len(numbers))
		for i, n := range numbers {
			vals[i] = cty.NumberUIntVal(n)
		}
		return cty.ListVal(vals), nil
	},
})

// parseSqidOptions extracts sqids.Options from an optional cty object argument.
// When opts is cty.NilVal or null, returns Options{} with nil Blocklist, which
// causes sqids to apply the default blocklist via its internal filterBlocklist logic.
// To disable the blocklist, pass {blocklist: []}.
func parseSqidOptions(opts cty.Value) (sqids.Options, error) {
	o := sqids.Options{}

	if opts == cty.NilVal || opts.IsNull() {
		return o, nil
	}

	ty := opts.Type()
	if !ty.IsObjectType() {
		return o, fmt.Errorf("sqid options must be an object, got %s", ty.FriendlyName())
	}

	if ty.HasAttribute("alphabet") {
		v := opts.GetAttr("alphabet")
		if !v.IsNull() {
			if v.Type() != cty.String {
				return o, fmt.Errorf("sqid options.alphabet must be a string")
			}
			o.Alphabet = v.AsString()
		}
	}

	if ty.HasAttribute("min_length") {
		v := opts.GetAttr("min_length")
		if !v.IsNull() {
			if v.Type() != cty.Number {
				return o, fmt.Errorf("sqid options.min_length must be a number")
			}
			bf := v.AsBigFloat()
			if !bf.IsInt() {
				return o, fmt.Errorf("sqid options.min_length must be an integer")
			}
			n, _ := bf.Int64()
			if n < 0 || n > 255 {
				return o, fmt.Errorf("sqid options.min_length must be between 0 and 255, got %d", n)
			}
			o.MinLength = uint8(n)
		}
	}

	if ty.HasAttribute("blocklist") {
		v := opts.GetAttr("blocklist")
		if !v.IsNull() {
			if !v.Type().IsListType() || v.Type().ElementType() != cty.String {
				return o, fmt.Errorf("sqid options.blocklist must be a list of strings")
			}
			// Explicitly set to non-nil empty slice so sqids skips the default blocklist.
			o.Blocklist = []string{}
			for it := v.ElementIterator(); it.Next(); {
				_, elem := it.Element()
				o.Blocklist = append(o.Blocklist, elem.AsString())
			}
		}
	}

	return o, nil
}

// ctyNumberToUint64 extracts a non-negative integer from a cty.Number value.
func ctyNumberToUint64(v cty.Value, funcName string) (uint64, error) {
	bf := v.AsBigFloat()
	if bf.Sign() < 0 {
		return 0, fmt.Errorf("%s: numbers must be non-negative, got %s", funcName, bf.Text('f', 0))
	}
	if !bf.IsInt() {
		return 0, fmt.Errorf("%s: numbers must be integers, got %s", funcName, bf.Text('f', -1))
	}
	n, accuracy := bf.Uint64()
	if accuracy != big.Exact {
		return 0, fmt.Errorf("%s: number out of range for uint64", funcName)
	}
	return n, nil
}
