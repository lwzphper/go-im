package config

import (
	"go-im/pkg/config"
	"go-im/pkg/logger"
	"go.uber.org/zap"
)

type FileBuilder struct{}

func NewFileBuilder() *FileBuilder {
	return &FileBuilder{}
}

func (f *FileBuilder) Builder(path string) {
	if err := config.LoadConfigFile(path, C); err != nil {
		logger.Error("load config error", zap.Error(err))
		panic(err)
	}

	builderCommon()
}
