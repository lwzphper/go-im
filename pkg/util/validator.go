package util

import (
	"github.com/gin-gonic/gin"
	"go-im/pkg/logger"
	"go-im/pkg/response"
	"reflect"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

const (
	EnLocale = "en"
	ZhLocale = "zh"
)

var Translator ut.Translator

// HandleValidatorError 处理表单验证错误
func HandleValidatorError(c *gin.Context, err error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		response.FormValidError(c.Writer, "[err] 请求参数有误")
		return
	}

	// 多条错误信息，只显示第一条
	for _, val := range errs.Translate(Translator) {
		response.FormValidError(c.Writer, val)
		return
	}
}

// Validator 初始验证器
func Validator(locale string) {
	// 修改gin框架中的validator引擎属性, 实现定制
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	// 注册一个获取 label 的tag的自定义方法
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := field.Tag.Get("label")
		if name == "" {
			return field.Name
		}

		return name
	})

	// 设置语言
	setLocale(v, locale)
}

// 设置语言
func setLocale(v *validator.Validate, locale string) {
	zhT := zh.New() // 中文翻译器
	enT := en.New() // 英文翻译器
	// 第一个参数是备用的语言环境，后面的参数是应该支持的语言环境
	uni := ut.New(enT, zhT, enT)
	var ok bool

	if Translator, ok = uni.GetTranslator(locale); !ok {
		logger.Errorf("uni.GetTranslator error:%s", locale)

		return
	}

	var err error

	switch locale {
	case ZhLocale:
		err = zh_translations.RegisterDefaultTranslations(v, Translator)
	default:
		err = en_translations.RegisterDefaultTranslations(v, Translator)
	}

	if err != nil {
		logger.Errorf("Register Default transition error:%v", err)
	}
}
