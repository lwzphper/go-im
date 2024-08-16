package mysql

import (
	"fmt"
	"github.com/pkg/errors"
	"go-im/pkg/file"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"os"
	"path/filepath"
	"time"
)

// MysqlConn mysql 连接信息
type MysqlConn struct {
	Host     string
	Port     string
	UserName string
	Password string
}

type MysqlConfig struct {
	ClientName  string
	Host        string
	Port        string
	UserName    string
	Password    string
	Database    string
	MaxOpenConn int
	MaxIdleConn int
	MaxLifeTime int
	MaxIdleTime int
	TablePrefix string
	LogFileName string
	LogLevel    logger.LogLevel
	Sources     []MysqlConn
	Replicas    []MysqlConn
}

func NewMysql(cfg *MysqlConfig) *Mysql {
	return &Mysql{
		MysqlConfig: cfg,
	}
}

type Mysql struct {
	*MysqlConfig

	// sqlDB *sql.DB
	// lock  sync.Mutex
}

// GetClient 获取 gorm 对象
func (m *Mysql) GetClient() *gorm.DB {
	client := GetMysqlClient(m.ClientName)
	if client == nil {
		panic("数据库未初始化")
	}

	return client.DB
}

func (m *Mysql) InitDB() error {
	options := []Option{
		WithMaxOpenConn(m.MaxOpenConn),
		WithMaxIdleConn(m.MaxIdleConn),
		WithConnMaxLifeSecond(time.Duration(m.MaxLifeTime)),
		WithMaxIdleTime(time.Duration(m.MaxIdleTime)),
		WithSources(m.getMasterSlaveInfo(m.Sources)),
		WithReplicas(m.getMasterSlaveInfo(m.Replicas)),
	}

	// 设置日志
	logConf := NewDefaultLoggerConf()
	if m.LogLevel == 0 {
		logConf.LogLevel = logger.Warn
	}

	logConf.LogLevel = m.LogLevel
	var logWriter io.Writer

	if m.LogFileName == "" {
		logWriter = os.Stdout
	} else {
		err := file.IsNotExistMkDir(filepath.Dir(m.LogFileName))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("cannot create dir：%s", m.LogFileName))
		}

		open, err := os.OpenFile(m.LogFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrap(err, "cannot open mysql log file")
		}
		logWriter = open
	}
	options = append(options, WithLogger(logWriter, logConf))

	// 初始化客户端
	if err := InitMysqlClient(m.ClientName, m.UserName, m.Password, m.Host, m.Port, m.Database, options...); err != nil {
		return err
	}

	return nil
}

// 获取主从连接信息
func (m *Mysql) getMasterSlaveInfo(conn []MysqlConn) []DBConnInfo {
	result := make([]DBConnInfo, 0)
	for _, item := range conn {
		result = append(result, DBConnInfo{
			Host:     item.Host,
			Port:     item.Port,
			Username: item.UserName,
			Password: item.Password,
		})
	}

	return result
}
