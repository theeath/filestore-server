package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)
var db *sql.DB
//数据库连接
func init()  {
	db,_ = sql.Open("mysql","root:123456@tcp(127.0.0.1:3308)/fileserver?charset=utf8")

	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil{
		fmt.Printf("failed to connect to mysql err :%s",err.Error())
		//os.Exit(1)
	}
	//defer db.Close()
}
func DBConn() *sql.DB {
	return db
}
func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		//将行数据保存到record字典
		err := rows.Scan(scanArgs...)
		checkErr(err)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}