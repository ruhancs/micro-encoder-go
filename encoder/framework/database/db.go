package database

import (
	"encoder/domain"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/lib/pq"
)

type Database struct {
	Db *gorm.DB
	Dsn string //string de conexao
	DsnTest string
	DbType string
	DbtypeTest string
	Debug bool
	AutoMigrateDb bool
	Env string
}

func NewDb() *Database {
	return &Database{}
}

func NewDbTest() *gorm.DB {
	dbInstance := NewDb()
	dbInstance.Env = "test"
	dbInstance.DbtypeTest = "sqlite3"
	dbInstance.DsnTest = ":memory:"
	dbInstance.AutoMigrateDb = true
	dbInstance.Debug = true

	connection,err := dbInstance.Connect()
	if err != nil {
		log.Fatalf("Test DB Error: %v", err)
	}

	return connection
}

func (d *Database) Connect() (*gorm.DB, error) {
	var err error
	if d.Env != "test" {
		d.Db, err = gorm.Open(d.DbType, d.Dsn)
		//d.Db.Raw("PRAGMA foreign_keys=ON")
	} else {
		d.Db, err = gorm.Open(d.DbtypeTest,d.DsnTest)
	}
	if err != nil {
		return nil, err
	}

	if d.Debug {
		d.Db.LogMode(true)//aparecer todos logs do db
	}

	if d.AutoMigrateDb {
		//realizar as migracoes das entidades
		d.Db.AutoMigrate(&domain.Video{}, &domain.Job{})
		//video_id em Job se relaciona na tabela videos (id)
		d.Db.Model(domain.Job{}).AddForeignKey("video_id", "videos (id)", "CASCADE", "CASCADE")
	}

	return d.Db, nil
}