package mysql

import (
	"context"
	"go-im/pkg/logger"
	"reflect"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// 事务key
type ctxTransactionKey struct{}

type Options func(*gorm.DB) *gorm.DB

func CtxWithTransaction(c context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(c, ctxTransactionKey{}, tx)
}

// WithTableName 设置表名（跨库事务时使用）。如：数据库名.表名
func WithTableName(name string) Options {
	return func(db *gorm.DB) *gorm.DB {
		return db.Table(name)
	}
}

func GetDB(ctx context.Context, defaultDB *gorm.DB, opts ...Options) *gorm.DB {
	fromCtx, err := GetTransactionFromCtx(ctx)
	if err != nil {
		logger.Panicf("get transition from ctx error：%v", err)

		return nil
	}

	var db *gorm.DB
	if fromCtx != nil { // 事务操作
		db = fromCtx
	} else {
		db = defaultDB.WithContext(ctx)
	}

	for _, opt := range opts {
		db = opt(db)
	}

	return db
}

func GetTransactionFromCtx(c context.Context) (*gorm.DB, error) {
	t := c.Value(ctxTransactionKey{})
	if t != nil {
		tx, ok := t.(*gorm.DB)
		if !ok {
			return nil, errors.Errorf("unexpect context value type: %s", reflect.TypeOf(tx))
		}
		return tx, nil
	}
	return nil, nil
}

// Transaction 事务操作
func Transaction(ctx context.Context, db *gorm.DB, fc func(txctx context.Context) error) error {
	db = db.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		txctx := CtxWithTransaction(ctx, tx)

		return fc(txctx)
	})
}
