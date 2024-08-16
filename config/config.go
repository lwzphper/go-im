package config

import (
	"fmt"
	"go-im/pkg/logger"
	"go-im/pkg/mysql"
	"go-im/pkg/redis"
	util "go-im/pkg/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormLogger "gorm.io/gorm/logger"
	"io"
	"os"
)

var C = NewConfig()

func NewConfig() *Config {
	return &Config{
		App: App{
			Name:        "go-im",
			Env:         "debug",
			InDocker:    true,
			GatewayAddr: ":9001",
		},
		Consul: Consul{
			Host:                "127.0.0.1:8500",
			Tag:                 "go-im",
			CheckTimeout:        "5s",
			CheckInterval:       "10s",
			CheckDeadDeregister: "30s",
		},
		Jwt: Jwt{
			Secret: "RMWzbVKDUI1SuynlMBn",
			TTL:    600,
		},
		Logging: Logging{
			Level: zap.DebugLevel.String(),
		},
		Redis: Redis{
			Name:        "default",
			Addr:        "127.0.0.1:6379",
			Password:    "",
			Database:    0,
			MinIdleConn: 1,
			PoolSize:    32,
			MaxRetries:  5,
		},
		Mysql: []Mysql{
			{
				ClientName: mysql.DefaultClient,
				Host:       "127.0.0.1",
				Port:       "3306",
				UserName:   "root",
				Password:   "123456",
				Database:   "go_im",
			},
		},
	}
}

type Config struct {
	App     App     `toml:"app" yaml:"app" mapstructure:"app" env:"APP"`
	Consul  Consul  `toml:"consul" yaml:"consul" mapstructure:"consul" env:"APP"`
	Jwt     Jwt     `toml:"jwt" yaml:"jwt" mapstructure:"jwt"`
	Logging Logging `toml:"logging" yaml:"logging" mapstructure:"logging"`
	Redis   Redis   `toml:"redis" yaml:"redis" mapstructure:"redis"`
	Mysql   []Mysql `toml:"mysql" yaml:"mysql" mapstructure:"mysql"`
}

// GetGatewayHost 获取网关地址
func (c *Config) GetGatewayHost() string {
	address, err := util.SplitAddress(c.App.GatewayAddr, false)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("http://%s", address.String())
}

// GetGatewayWsAddr 获取网关ws地址
func (c *Config) GetGatewayWsAddr() string {
	address, err := util.SplitAddress(c.App.GatewayAddr, c.App.InDocker)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("ws://%s/proxy", address.String())
}

// App 项目配置
type App struct {
	Name        string `toml:"name" yaml:"name" mapstructure:"name" env:"APP_NAME"`
	Env         string `toml:"env" yaml:"env" mapstructure:"env" env:"APP_ENV"`
	InDocker    bool   `toml:"in_docker" yaml:"in_docker" mapstructure:"in_docker" env:"APP_IN_DOCKER"` // 项目是否允许在 docker 环境中
	GatewayAddr string `toml:"gateway_addr" yaml:"gateway_addr" mapstructure:"gateway_addr" env:"GATEWAY_ADDR"`
}

// Consul consul配置
type Consul struct {
	Host                string `toml:"host" yaml:"host" mapstructure:"host" env:"CONSUL_HOST"`
	Tag                 string `toml:"tag" yaml:"tag" mapstructure:"tag" env:"CONSUL_TAG"`
	CheckTimeout        string `toml:"check_timeout" yaml:"check_timeout" mapstructure:"check_timeout" env:"CONSUL_CHECK_TIMEOUT"`
	CheckInterval       string `toml:"check_interval" yaml:"check_interval" mapstructure:"check_interval" env:"CONSUL_CHECK_INTERVAL"`
	CheckDeadDeregister string `toml:"check_dead_deregister" yaml:"check_dead_deregister" mapstructure:"check_dead_deregister" env:"CONSUL_CHECK_DEAD_DEREGISTER"`
}

type Mysql struct {
	ClientName  string              `toml:"client_name" yaml:"client_name" mapstructure:"client_name" env:"MYSQL_CLIENT_NAME"`
	Host        string              `toml:"host" yaml:"host" mapstructure:"host" env:"MYSQL_HOST"`
	Port        string              `toml:"port" yaml:"port" mapstructure:"port" env:"MYSQL_PORT"`
	UserName    string              `toml:"username" yaml:"username" mapstructure:"username" env:"MYSQL_USERNAME"`
	Password    string              `toml:"password" yaml:"password" mapstructure:"password" env:"MYSQL_PASSWORD"`
	Database    string              `toml:"database" yaml:"database" mapstructure:"database" env:"MYSQL_DATABASE"`
	MaxOpenConn int                 `toml:"max_open_conn" yaml:"max_open_conn" mapstructure:"max_open_conn" env:"MYSQL_MAX_OPEN_CONN"`
	MaxIdleConn int                 `toml:"max_idle_conn" yaml:"max_idle_conn" mapstructure:"max_idle_conn" env:"MYSQL_MAX_IDLE_CONN"`
	MaxLifeTime int                 `toml:"max_life_time" yaml:"max_life_time" mapstructure:"max_life_time" env:"MYSQL_MAX_LIFE_TIME"`
	MaxIdleTime int                 `toml:"max_idle_time" yaml:"max_idle_time" mapstructure:"max_idle_time" env:"MYSQL_MAX_IDLE_TIME"`
	TablePrefix string              `toml:"table_prefix" yaml:"table_prefix" mapstructure:"table_prefix" env:"MYSQL_TABLE_PREFIX"`
	LogFileName string              `toml:"log_file_name" yaml:"log_file_name" mapstructure:"log_file_name" env:"MYSQL_LOG_FILE_NAME"`
	LogLevel    gormLogger.LogLevel `toml:"log_level" yaml:"log_level" mapstructure:"log_level" env:"MYSQL_LOG_LEVEL"`
	Sources     []MysqlConn         `toml:"sources" yaml:"sources" mapstructure:"sources" env:"MYSQL_SOURCES"`     // 主库配置
	Replicas    []MysqlConn         `toml:"replicas" yaml:"replicas" mapstructure:"replicas" env:"MYSQL_REPLICAS"` // 从库配置
}

func (m *Mysql) Config() *mysql.MysqlConfig {
	sourceConns := make([]mysql.MysqlConn, len(m.Sources))
	for _, conn := range m.Sources {
		sourceConns = append(sourceConns, conn.config())
	}

	replicasConns := make([]mysql.MysqlConn, len(m.Replicas))
	for _, conn := range m.Replicas {
		replicasConns = append(replicasConns, conn.config())
	}

	return &mysql.MysqlConfig{
		ClientName:  m.ClientName,
		Host:        m.Host,
		Port:        m.Port,
		UserName:    m.UserName,
		Password:    m.Password,
		Database:    m.Database,
		MaxOpenConn: m.MaxOpenConn,
		MaxIdleConn: m.MaxIdleConn,
		MaxLifeTime: m.MaxLifeTime,
		MaxIdleTime: m.MaxIdleTime,
		TablePrefix: m.TablePrefix,
		LogFileName: m.LogFileName,
		LogLevel:    m.LogLevel,
		Sources:     sourceConns,
		Replicas:    replicasConns,
	}
}

// MysqlConn mysql 连接信息
type MysqlConn struct {
	Host     string `toml:"host" yaml:"host" mapstructure:"host" env:"MYSQL_HOST"`
	Port     string `toml:"port" yaml:"port" mapstructure:"port" env:"MYSQL_PORT"`
	UserName string `toml:"username" yaml:"username" mapstructure:"username" env:"MYSQL_USERNAME"`
	Password string `toml:"password" yaml:"password" mapstructure:"password" env:"MYSQL_PASSWORD"`
}

func (c *MysqlConn) config() mysql.MysqlConn {
	return mysql.MysqlConn{
		Host:     c.Host,
		Port:     c.Port,
		UserName: c.UserName,
		Password: c.Password,
	}
}

type Jwt struct {
	Secret string `toml:"secret" yaml:"secret" mapstructure:"secret" env:"JWT_SECRET"`
	TTL    int64  `toml:"ttl" yaml:"ttl" mapstructure:"ttl" env:"JWT_TTL"`
}

// Logging 日志
type Logging struct {
	Name     string `toml:"name" yaml:"name" mapstructure:"name" env:"LOGGING_NAME"` // 配置唯一标识
	FileName string `toml:"file_name" yaml:"file_name" mapstructure:"file_name" env:"LOGGING_FILE_NAME"`
	Level    string `toml:"level" yaml:"level" mapstructure:"level" env:"LOGGING_LEVEL"`
}

type Redis struct {
	Name        string `toml:"name" yaml:"name" mapstructure:"name" env:"REDIS_NAME"` // 配置唯一标识
	Addr        string `toml:"addr" yaml:"addr" mapstructure:"addr" env:"REDIS_ADDR"`
	Password    string `toml:"password" yaml:"password" mapstructure:"password" env:"REDIS_PASSWORD"`
	Database    int    `toml:"database" yaml:"database" mapstructure:"database" env:"REDIS_DATABASE"`
	MinIdleConn int    `toml:"min_idle_conn" yaml:"min_idle_conn" mapstructure:"min_idle_conn" env:"REDIS_MIN_IDLE_CONN"`
	PoolSize    int    `toml:"pool_size" yaml:"pool_size" mapstructure:"pool_size" env:"REDIS_POOL_SIZE"`
	MaxRetries  int    `toml:"max_retries" yaml:"max_retries" mapstructure:"max_retries" env:"REDIS_MAX_RETRIES"`
}

func (r *Redis) Config() redis.Config {
	return redis.Config{
		Name:        r.Name,
		Addr:        r.Addr,
		Password:    r.Password,
		Database:    r.Database,
		MinIdleConn: r.MinIdleConn,
		PoolSize:    r.PoolSize,
		MaxRetries:  r.MaxRetries,
	}
}

func builderCommon() {
	// 设置日志
	setLogger(C.Logging.Name, C.Logging.Level)
	// 初始化 redis
	redis.NewRedis(C.Redis.Config()).Init()
	// 加载 mysql
	for _, conf := range C.Mysql {
		if err := mysql.NewMysql(conf.Config()).InitDB(); err != nil {
			panic(err)
		}
	}
}

// 设置日志
func setLogger(path, level string) {
	var (
		writer io.Writer
		err    error
	)

	if path == "" {
		writer = os.Stderr
	} else {
		writer, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(fmt.Sprintf("cannot open log file: %v", err))
		}
	}

	lv := zapcore.InfoLevel
	if level != "" {
		if lv, err = zapcore.ParseLevel(level); err != nil {
			lv = zapcore.InfoLevel // 如果解析出错，设置 info 级别
		}
	}

	l := logger.New(writer, lv)
	logger.ResetDefault(l)
}
