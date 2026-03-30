# sqid-cty-funcs

A Go module providing [sqids](https://sqids.org) encode/decode functions for use in [go-cty](https://github.com/zclconf/go-cty) / HCL2 evaluation contexts.

Sqids are short, URL-safe IDs generated from numbers. They are reversible and can encode one or more non-negative integers into a compact string.

## Installation

```
go get github.com/tsarna/sqid-cty-funcs
```

## Usage

```go
import (
    sqidcty "github.com/tsarna/sqid-cty-funcs"
    "github.com/zclconf/go-cty/cty/function"
)

// Register all functions in an HCL eval context
funcs := sqidcty.GetSqidFunctions()
// funcs is map[string]function.Function — merge into your eval context
```

## Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `sqid` | `sqid(id number\|list(number)[, options object]) string` | Encodes one or more non-negative integers into a sqid string |
| `unsqid` | `unsqid(s string[, options object]) list(number)` | Decodes a sqid string into a list of non-negative integers |

### `sqid(id[, options])`

Encodes a single non-negative integer or a list of non-negative integers into a sqid string. The encoding is reversible: the same numbers always produce the same sqid (for a given options configuration).

```hcl
id  = sqid(42)          # single number → e.g. "MhPE"
ids = sqid([1, 2, 3])   # list of numbers → e.g. "86Rf07"
```

An empty list encodes to an empty string.

### `unsqid(s[, options])`

Decodes a sqid string back into a list of non-negative integers. Returns an empty list for an empty string or any input that cannot be decoded (invalid characters, tampered IDs). This function never returns an error — invalid input silently produces an empty list.

```hcl
nums = unsqid("86Rf07")   # → [1, 2, 3]
nums = unsqid("")          # → []
nums = unsqid("???")       # → [] (invalid, no error)
```

### Options

Both functions accept an optional second argument — an object with any of the following attributes:

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `alphabet` | string | 62-char alphanumeric | Custom character set for encoding. Must have at least 3 unique ASCII characters. |
| `min_length` | number | `0` | Minimum length of the generated ID. IDs shorter than this are padded. Range: 0–255. |
| `blocklist` | list(string) | default blocklist | Words to exclude from generated IDs. Pass `[]` to disable the default blocklist entirely. Omitting this attribute (or setting it to `null`) keeps the default blocklist active. |

```hcl
# Custom alphabet
id = sqid(1, { alphabet = "abcdefghijklmnopqrstuvwxyz0123456789" })

# Minimum length of 10
id = sqid(1, { min_length = 10 })

# No profanity filtering
id = sqid(1, { blocklist = [] })

# Combine options
id = sqid(1, { min_length = 8, alphabet = "abcdefghijklmnopqrstuvwxyz0123456789" })
```

## Notes

- Numbers must be non-negative integers. Floats or negative values produce an error.
- `unsqid` is the inverse of `sqid`: `unsqid(sqid(nums)) == nums` for the same options.
- The default blocklist filters common profanity in multiple languages. Using a custom alphabet automatically filters the blocklist to only words that can be formed from that alphabet.
