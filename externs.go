package sqidcty

import _ "embed"

//go:embed externs.cty
var externsCty []byte

// ExternsFilename is the name reported for the embedded declarations in
// diagnostics.
const ExternsFilename = "sqid-cty-funcs/externs.cty"

// Externs returns the functy `//functy:extern` declarations for the sqid functions:
// their real signatures, which their cty metadata cannot express.
//
// sqid()'s first argument is a union — a single number, or a list of numbers — and cty
// has no union type, so its metadata can only say "dynamic"; it is declared here as one
// form per arm. And both functions take an optional trailing options object, which cty
// can only fake with a variadic, leaving it shapeless — nothing in the metadata says it
// holds alphabet, min_length, and blocklist. The declarations spell it out.
//
// The bytes are opaque to this package: it does not import functy, and nothing here
// parses them. A functy host registers them:
//
//	parser.RegisterExterns(sqidcty.Externs(), sqidcty.ExternsFilename)
func Externs() []byte { return externsCty }
