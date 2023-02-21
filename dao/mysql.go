package dao

import (
	"QTransaction/config"
	"fmt"
	"github.com/beego/beego/v2/adapter/logs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// db 全局MySQL数据库操作对象
var db *gorm.DB

//InitMysql 链接数据库
func InitMysql() {
	cfg := config.Config().Mysql
	if cfg.Ip == "" {
		panic("invalid mysql ip")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Ip, cfg.Port, cfg.DbName)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	CreateMysql() //初始化建表
}
func CreateMysql() {
	if err := db.AutoMigrate(
	//
	); err != nil {
		logs.Info(fmt.Printf("automigrate table error: %v \n", err))
	}
}
