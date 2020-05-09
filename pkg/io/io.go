package io

import (
	"bytes"
	"fmt"
)

type Buffer struct{ bytes.Buffer }

func (b *Buffer) Writef(f string, a ...interface{}) {
	_, _ = b.WriteString(fmt.Sprintf(f, a...))
}
