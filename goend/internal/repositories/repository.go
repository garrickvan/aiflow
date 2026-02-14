package repositories

import (
	"fmt"
	"time"

	"aiflow/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Repository 数据库操作仓库
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建新的数据库仓库实例
// 优化配置：启用连接池、WAL模式、缓存等提升性能
func NewRepository(dbPath string) (*Repository, error) {
	// SQLite性能优化参数
	// _journal_mode=WAL: 启用WAL模式，提升并发读写性能
	// _busy_timeout=5000: 设置忙等待超时5秒
	// _cache_size=10000: 设置缓存页数，提升查询性能
	// _synchronous=NORMAL: 同步模式，平衡性能和数据安全
	// _temp_store=MEMORY: 临时表存储在内存中
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_cache_size=10000&_synchronous=NORMAL&_temp_store=MEMORY",
		dbPath)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层SQL连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 配置连接池参数
	sqlDB.SetMaxOpenConns(10)           // 最大打开连接数
	sqlDB.SetMaxIdleConns(5)            // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	// 自动迁移数据库表结构（必须先执行，确保表结构正确）
	err = db.AutoMigrate(
		&models.Skill{},
		&models.Tag{},
		&models.SkillTag{},
		&models.SkillToken{},
		&models.JobTask{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// 执行数据迁移（处理旧表结构变更，在AutoMigrate之后）
	if err := models.MigrateData(db); err != nil {
		return nil, fmt.Errorf("failed to migrate data: %w", err)
	}

	// 创建数据库索引优化查询性能
	if err := models.CreateIndexes(db); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return &Repository{db: db}, nil
}

// GetDB 获取数据库连接
func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

// NewEmptyRepository 创建一个空的Repository实例（用于数据库初始化失败时）
// 返回的Repository的db字段为nil，调用GetDB()会返回nil
func NewEmptyRepository() *Repository {
	return &Repository{db: nil}
}
