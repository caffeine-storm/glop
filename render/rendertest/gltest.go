package rendertest

type GlTestBuilder struct{}

func (b *GlTestBuilder) Run(fn func()) {
	WithGl(fn)
}

func GlTest() *GlTestBuilder {
	return &GlTestBuilder{}
}
