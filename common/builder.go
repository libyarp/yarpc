package common

import (
	"fmt"
	"strings"
)

type Builder struct {
	b *strings.Builder
}

func (b *Builder) Printf(m string, a ...any) {
	b.b.WriteString(fmt.Sprintf(m, a...))
}

func (b *Builder) Printfln(m string, a ...any) {
	b.Printf(m+"\n", a...)
}

func (b *Builder) String() string {
	return b.b.String()
}

func (b *Builder) Absorb(o *Builder) {
	b.b.WriteString(o.String())
}

func NewBuilder() *Builder {
	return &Builder{b: &strings.Builder{}}
}
