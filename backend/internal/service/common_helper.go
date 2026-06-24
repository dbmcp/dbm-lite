package service

import "strings"

func camelToSnakeHelper(s string) string {
var b strings.Builder
for i, ch := range s {
if ch >= 'A' && ch <= 'Z' {
if i > 0 {
b.WriteByte('_')
}
b.WriteByte(byte(ch) + 32)
} else {
b.WriteRune(ch)
}
}
return b.String()
}
