package repo

import (
	"context"
	"github.com/pkg/errors"
	"go-im/internal/logic/user"
	"go-im/internal/logic/user/model"
	"go-im/pkg/mysql"
	"go-im/pkg/util"
	"gorm.io/gorm"
)

func NewUserRepo() *UserRepo {
	return &UserRepo{
		db: mysql.GetMysqlClient(mysql.DefaultClient),
	}
}

type UserRepo struct {
	db *mysql.DB
}

// Add 保存用户信息
func (d *UserRepo) Add(ctx context.Context, userModel model.User) (uint64, error) {
	err := d.db.DB.Create(&userModel).Error
	if err != nil {
		util.LogError(ctx, err)
		return 0, user.ErrDBOperate
	}
	return userModel.Id, nil
}

// GetByUsername 根据账号获取用户信息
func (d *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var userModel model.User
	err := d.db.DB.First(&userModel, "username = ?", username).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		util.LogError(ctx, err)
		return nil, user.ErrDBOperate
	}
	return &userModel, nil
}

// GetById 通过id获取用户信息
func (d *UserRepo) GetById(ctx context.Context, id uint64) (*model.User, error) {
	var userModel model.User
	err := d.db.DB.First(&userModel, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		util.LogError(ctx, err)
		return nil, user.ErrDBOperate
	}
	return &userModel, nil
}
