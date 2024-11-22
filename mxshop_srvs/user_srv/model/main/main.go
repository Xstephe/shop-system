package main

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/anaskhan96/go-password-encoder"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"mxshop_srvs/user_srv/model"
)

func genMd5(code string) string {
	Md5 := md5.New()
	_, _ = io.WriteString(Md5, code)
	return hex.EncodeToString(Md5.Sum(nil))
}

func main() {
	dsn := "root:513513@tcp(192.168.182.130:3306)/mxshop_user_srv?charset=utf8mb4&parseTime=true&loc=Local"
	//设置全局的logger，打印出sql语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  true,        // Disable color
		},
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}

	//_ = db.AutoMigrate(&model.User{})
	options := &password.Options{10, 100, 32, sha512.New}
	salt, encodedPwd := password.Encode("admin123", options)
	NewPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	for i := 0; i < 10; i++ {
		user := &model.User{
			Nickname: fmt.Sprintf("bobby%d", i),
			Mobile:   fmt.Sprintf("1853985732%d", i),
			Password: NewPassword,
		}
		db.Save(&user)
	}
}
