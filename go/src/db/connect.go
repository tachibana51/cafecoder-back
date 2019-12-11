package cafedb

import(
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
)
//todo impl
func NewCon() {
	db, err := sql.Open("mysql", "root:root@/kakecoder")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
}