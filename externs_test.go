package sqidcty

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

// externDeclRE matches a top-level declaration in externs.cty. The file is parsed here
// with a regex rather than with functy on purpose: this package must not depend on functy
// (its bytes are opaque to it), and the check only needs the name set. A name may appear
// on several lines — sqid is an overload set — so the result is a set.
var externDeclRE = regexp.MustCompile(`(?m)^func (\w+)\(`)

// TestExternsCoverEveryFunction is the drift guard. Both functions take an optional
// options object cty renders shapeless, and sqid additionally a number-or-list union cty
// cannot name — so both are declared. Adding a function without a declaration, or
// declaring one that no longer exists, fails here.
func TestExternsCoverEveryFunction(t *testing.T) {
	declared := make(map[string]bool)
	for _, m := range externDeclRE.FindAllStringSubmatch(string(Externs()), -1) {
		declared[m[1]] = true
	}

	funcs := GetSqidFunctions()
	for name := range funcs {
		assert.True(t, declared[name],
			"%s() is provided by GetSqidFunctions but has no declaration in externs.cty", name)
	}
	for name := range declared {
		assert.Contains(t, funcs, name,
			"externs.cty declares %s(), which GetSqidFunctions does not provide", name)
	}
}

// The bytes must declare themselves an extern file: functy's RegisterExterns verifies the
// directive rather than forcing the mode, so that this same file is a valid standalone
// .cty that `functy fmt` and `functy symbols` can open.
func TestExternsCarryTheDirective(t *testing.T) {
	require.True(t, strings.HasPrefix(string(Externs()), "//functy:extern\n"),
		"externs.cty must begin with the //functy:extern directive")
}

// Every function and every parameter carries a cty description. The metadata is the only
// documentation a non-functy cty host can see, and the only thing functy's own doc()
// reads (doc() does not consult the extern).
func TestEverythingIsDescribed(t *testing.T) {
	for name, fn := range GetSqidFunctions() {
		assert.NotEmpty(t, fn.Description(), "%s() has no cty Description", name)

		for _, p := range fn.Params() {
			assert.NotEmpty(t, p.Description, "%s() parameter %q has no Description", name, p.Name)
		}
		if vp := fn.VarParam(); vp != nil {
			assert.NotEmpty(t, vp.Description, "%s() variadic parameter %q has no Description", name, vp.Name)
		}
	}
}

// The options variadic exists to fake an *optional* argument, not a repeatable one. A cty
// VarParam has no upper bound of its own, so without an explicit ceiling both functions
// would accept and silently ignore extra arguments.
func TestExcessArgumentsAreRejected(t *testing.T) {
	opts := cty.EmptyObjectVal
	for name, call := range map[string]func() error{
		"sqid": func() error {
			_, err := SqidFunc.Call([]cty.Value{cty.NumberIntVal(1), opts, opts})
			return err
		},
		"unsqid": func() error {
			_, err := UnsqidFunc.Call([]cty.Value{cty.StringVal("abc"), opts, opts})
			return err
		},
	} {
		assert.Error(t, call(), "%s() silently accepted a surplus argument", name)
	}
}
