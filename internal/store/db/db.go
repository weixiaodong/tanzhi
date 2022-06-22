package db

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/weixiaodong/tanzhi/component/log"
	"github.com/weixiaodong/tanzhi/config"
)

var (
	// db配置的job任务
	schema1 = `
		CREATE TABLE IF NOT EXISTS t_job (
		  id INTEGER PRIMARY KEY AUTOINCREMENT,
		  name varchar(64) NOT NULL,
		  expr varchar(32) NOT NULL,
		  command varchar(128) NOT NULL DEFAULT '',
		  create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	// job执行结果记录
	schema2 = `
		CREATE TABLE IF NOT EXISTS t_job_record (
		  id INTEGER PRIMARY KEY AUTOINCREMENT,
		  name varchar(64) NOT NULL,
		  type varchar(32) NOT NULL,
		  command varchar(128) NOT NULL DEFAULT '',
		  result text NOT NULL DEFAULT '',
		  failed_cnt INT(10) NOT NULL DEFAULT '0',
		  create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  started_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  finished_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
)
var db *sqlx.DB

func Init() {
	var err error
	db, err = sqlx.Connect("sqlite3", "__job.db")
	if err != nil {
		log.Fatal(context.Background(), "init_db_failed", "err", err)
	}
	db.MustExec(schema1)
	db.MustExec(schema2)

	tx := db.MustBegin()
	command1 := &config.CommandConfig{
		Type:   "http",
		Method: "GET",
		Target: "https://www.qq.com",
	}
	command2 := &config.CommandConfig{
		Type:   "http",
		Method: "GET",
		Target: "https://cn.bing.com",
	}
	tx.MustExec("INSERT OR IGNORE INTO t_job (id, name, expr, command) VALUES ($1, $2, $3, $4)", 1, "db-job-1", "0 */3 * * * *", command1.Encode())
	tx.MustExec("INSERT OR IGNORE INTO t_job (id, name, expr, command) VALUES ($1, $2, $3, $4)", 2, "db-job-2", "0 */4 * * * *", command2.Encode())
	tx.Commit()

	return
}

type Job struct {
	Id         int       `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Expr       string    `db:"expr" json:"expr"`
	Command    string    `db:"command" json:"command"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
}

func GetJob(ctx context.Context) (res []*Job, err error) {
	sql1 := `
		SELECT *
		FROM t_job
		ORDER BY create_time desc
	`
	err = db.Select(&res, sql1)
	return
}
