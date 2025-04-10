# Search bar query syntax parser

Handle search query syntax with a basic parse tree

Examples ->

- `foo bar`
- `foo=1 bar=2`
- `foo=1 AND bar=2`
- `foo=1 OR bar=2`
- `(foo=1 OR bar=2) baz=3`

## Expression

- `and`, `AND` -- binary expression
- `or`, `OR` -- binary expression
- `NOT` -- unary expression

## Operators

The logic for these doesn't exist, but these represent operators that are pulled into the AST

- equality: `:`, `=`, `!=` (these are distinct, but should be equals)
- comparison: `<`, `>` `<=` `>=`
- like: `~`, `!~`
- `-` for negation of terms e.g. `-fish` (if you don't like fish) which should be the same as `NOT fish`

## Grouping

Parenentical expression `( foo=bar )` are parsed and put into the AST
