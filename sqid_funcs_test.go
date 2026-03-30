package sqidcty

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestSqidSingleNumber(t *testing.T) {
	result, err := SqidFunc.Call([]cty.Value{cty.NumberIntVal(1)})
	require.NoError(t, err)
	assert.Equal(t, cty.String, result.Type())
	assert.NotEmpty(t, result.AsString())
}

func TestSqidList(t *testing.T) {
	result, err := SqidFunc.Call([]cty.Value{
		cty.ListVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2), cty.NumberIntVal(3)}),
	})
	require.NoError(t, err)
	assert.Equal(t, cty.String, result.Type())
	assert.NotEmpty(t, result.AsString())
}

func TestSqidZero(t *testing.T) {
	result, err := SqidFunc.Call([]cty.Value{cty.NumberIntVal(0)})
	require.NoError(t, err)
	assert.NotEmpty(t, result.AsString())
}

func TestSqidRoundTripSingle(t *testing.T) {
	encoded, err := SqidFunc.Call([]cty.Value{cty.NumberIntVal(42)})
	require.NoError(t, err)

	decoded, err := UnsqidFunc.Call([]cty.Value{encoded})
	require.NoError(t, err)

	var nums []uint64
	for it := decoded.ElementIterator(); it.Next(); {
		_, v := it.Element()
		n, _ := v.AsBigFloat().Uint64()
		nums = append(nums, n)
	}
	require.Len(t, nums, 1)
	assert.Equal(t, uint64(42), nums[0])
}

func TestSqidRoundTripList(t *testing.T) {
	encoded, err := SqidFunc.Call([]cty.Value{
		cty.ListVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2), cty.NumberIntVal(3)}),
	})
	require.NoError(t, err)

	decoded, err := UnsqidFunc.Call([]cty.Value{encoded})
	require.NoError(t, err)

	var nums []uint64
	for it := decoded.ElementIterator(); it.Next(); {
		_, v := it.Element()
		n, _ := v.AsBigFloat().Uint64()
		nums = append(nums, n)
	}
	assert.Equal(t, []uint64{1, 2, 3}, nums)
}

func TestSqidMinLength(t *testing.T) {
	result, err := SqidFunc.Call([]cty.Value{
		cty.NumberIntVal(1),
		cty.ObjectVal(map[string]cty.Value{
			"min_length": cty.NumberIntVal(10),
		}),
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.AsString()), 10)
}

func TestSqidCustomAlphabet(t *testing.T) {
	alphabet := "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	opts := cty.ObjectVal(map[string]cty.Value{
		"alphabet": cty.StringVal(alphabet),
	})

	encoded, err := SqidFunc.Call([]cty.Value{cty.NumberIntVal(123), opts})
	require.NoError(t, err)

	decoded, err := UnsqidFunc.Call([]cty.Value{encoded, opts})
	require.NoError(t, err)

	var nums []uint64
	for it := decoded.ElementIterator(); it.Next(); {
		_, v := it.Element()
		n, _ := v.AsBigFloat().Uint64()
		nums = append(nums, n)
	}
	require.Len(t, nums, 1)
	assert.Equal(t, uint64(123), nums[0])
}

func TestSqidEmptyBlocklist(t *testing.T) {
	result, err := SqidFunc.Call([]cty.Value{
		cty.NumberIntVal(1),
		cty.ObjectVal(map[string]cty.Value{
			"blocklist": cty.ListValEmpty(cty.String),
		}),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.AsString())
}

func TestSqidEmptyOptions(t *testing.T) {
	result, err := SqidFunc.Call([]cty.Value{
		cty.NumberIntVal(1),
		cty.EmptyObjectVal,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.AsString())
}

func TestSqidNegativeNumber(t *testing.T) {
	_, err := SqidFunc.Call([]cty.Value{cty.NumberIntVal(-1)})
	assert.Error(t, err)
}

func TestSqidFloat(t *testing.T) {
	_, err := SqidFunc.Call([]cty.Value{cty.NumberFloatVal(1.5)})
	assert.Error(t, err)
}

func TestSqidWrongType(t *testing.T) {
	_, err := SqidFunc.Call([]cty.Value{cty.StringVal("hello")})
	assert.Error(t, err)
}

func TestSqidTooManyArgs(t *testing.T) {
	_, err := SqidFunc.Call([]cty.Value{
		cty.NumberIntVal(1),
		cty.EmptyObjectVal,
		cty.EmptyObjectVal,
	})
	assert.Error(t, err)
}

func TestSqidMinLengthOutOfRange(t *testing.T) {
	_, err := SqidFunc.Call([]cty.Value{
		cty.NumberIntVal(1),
		cty.ObjectVal(map[string]cty.Value{
			"min_length": cty.NumberIntVal(256),
		}),
	})
	assert.Error(t, err)
}

func TestUnsqidEmpty(t *testing.T) {
	result, err := UnsqidFunc.Call([]cty.Value{cty.StringVal("")})
	require.NoError(t, err)
	assert.Equal(t, cty.ListValEmpty(cty.Number), result)
}

func TestUnsqidInvalidChars(t *testing.T) {
	result, err := UnsqidFunc.Call([]cty.Value{cty.StringVal("!!!")})
	require.NoError(t, err)
	assert.Equal(t, cty.ListValEmpty(cty.Number), result)
}

func TestUnsqidTooManyArgs(t *testing.T) {
	_, err := UnsqidFunc.Call([]cty.Value{
		cty.StringVal("abc"),
		cty.EmptyObjectVal,
		cty.EmptyObjectVal,
	})
	assert.Error(t, err)
}

func TestGetSqidFunctions(t *testing.T) {
	funcs := GetSqidFunctions()
	assert.Contains(t, funcs, "sqid")
	assert.Contains(t, funcs, "unsqid")
	assert.NotZero(t, funcs["sqid"])
	assert.NotZero(t, funcs["unsqid"])
}
