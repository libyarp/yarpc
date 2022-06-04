package common

import (
	"fmt"
	"strings"
)

type Builder struct {
	buffer string
	level  int
}

func (b *Builder) Indent(fn func(*Builder)) {
	b.level++
	fn(b)
	b.level--
}

func (b *Builder) adjustIndent() {
	if b.level == 0 {
		return
	}

	if b.buffer[len(b.buffer)-1] == "\n"[0] {
		b.buffer += strings.Repeat("  ", b.level)
	}
}

func (b *Builder) Printf(m string, a ...any) {
	if len(m) != 0 {
		b.adjustIndent()
	}
	b.write(fmt.Sprintf(m, a...))
}

func (b *Builder) Printfln(m string, a ...any) {
	if len(m) != 0 {
		b.adjustIndent()
	}
	b.Printf(m+"\n", a...)
}

func (b *Builder) String() string {
	return b.buffer
}

func (b *Builder) Absorb(o *Builder) {
	b.write(o.String())
}

func (b *Builder) write(s string) {
	b.buffer += s
}

func (b *Builder) drop(size int) {
	b.buffer = b.buffer[0 : len(b.buffer)-size]
}

func NewBuilder() *Builder {
	return &Builder{}
}
