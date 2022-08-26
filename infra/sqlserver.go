package infra

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	DbPayment *sql.DB
	SqlConf   *DBData
)

type DBData struct {
	DB_DRIVER   string
	DB_USER     string
	DB_PASSWORD string
	DB_HOST     string
	DB_INSTANCE string
	DB_DATABASE string
	DB_ENCRYPT  string
}

func ConnectDB() *sql.DB {
	conection, err := sql.Open(SqlConf.DB_DRIVER, SqlConf.DB_DRIVER+"://"+SqlConf.DB_USER+":"+SqlConf.DB_PASSWORD+"@"+SqlConf.DB_HOST+"/"+SqlConf.DB_INSTANCE+"?"+"database="+SqlConf.DB_DATABASE+"&"+"encrypt="+SqlConf.DB_ENCRYPT+"")
	if err != nil {
		panic(err.Error())
	}
	return conection
}

func CheckDB() error {
	var err error
	ctx := context.Background()
	err = DbPayment.PingContext(ctx)
	if err != nil {
		return err
	}
	return err
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}
