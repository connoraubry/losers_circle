package db

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Options struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	Args     string
}

func DSNFromOptions(opts Options) string {
	var user string
	var password string
	var host string
	var port int
	var database string
	var args string

	if opts.User != "" {
		user = opts.User
	} else {
		user = "root"
	}

	if opts.Password != "" {
		password = opts.Password
	} else {
		password = "password"
	}

	if opts.Host != "" {
		host = opts.Host
	} else {
		host = "db"
	}

	if opts.Port != 0 {
		port = opts.Port
	} else {
		port = 3306
	}

	if opts.Database != "" {
		database = opts.Database
	} else {
		database = "circle"
	}

	if opts.Args != "" {
		args = opts.Args
	} else {
		args = "charset=utf8mb4&parseTime=True&loc=Local"
	}

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?%v",
		user, password, host, port, database, args)

	return dsn
}

func NewDB(opts Options) *gorm.DB {

	dsn := DSNFromOptions(opts)
	log.Info("Connecting to dsn", dsn)

	var db *gorm.DB
	var err error
	for tries := 3; tries > 0; tries-- {

		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.WithField("tries left", tries).Errorf("Error connecting to database: %v", err)

		if tries != 1 {
			time.Sleep(time.Second)
		}
	}
	if err != nil {
		log.Fatalf("Max connection tries reached, got error: %v", err)
	}

	log.Info("Connected to database")

	Migrate(db)
	return db
}

func Migrate(db *gorm.DB) {
	log.Info("Auto migrating database")
	db.AutoMigrate(&Year{})
	db.AutoMigrate(&Week{})
	db.AutoMigrate(&Team{})
	db.AutoMigrate(&Game{})
}
