package config

type DefaultBuilder struct{}

func NewDefault() *DefaultBuilder {
	return &DefaultBuilder{}
}

func (d *DefaultBuilder) Builder() {
	C = NewConfig()

	builderCommon()
}
