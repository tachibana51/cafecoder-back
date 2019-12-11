package cafedb

import(
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
	"../values"
	"regexp"
	"reflect"
)
//todo impl
type MyCon struct {
	DB *sql.DB
	Regex *regexp.Regexp
}

func NewCon(con *MyCon) (*MyCon) {
	db, err := sql.Open("mysql", values.MySQLDBN)
	if err != nil {
		panic(err.Error())
	}
	con.DB = db
	con.Regex = regexp.MustCompile(`[^(a-zA-Z\._-=@)]+`)
	return con
}

func SafeSelect(con *MyCon, sql string, bindData... interface{} ) (*sql.Rows, error) {
	//prepare
	for i, data := range bindData {
		//string assert
		if reflect.TypeOf(data) != reflect.TypeOf("") {
			continue
		}
		bindData[i] = con.Regex.ReplaceAllString(data.(string), "")
	}
	//bind
	bindSql := fmt.Sprintf(sql, bindData...)
	//execute
	rows, err := con.DB.Query(bindSql)
	if err != nil{
		return nil, err
	}
	return rows, err
}

func PrepareExec(con *MyCon, sql string, bindData... interface{}) (bool, error){
	stmt, err := con.DB.Prepare(sql)
	if err != nil {
		return false, err
	}
	stmt.Exec(bindData...)
	return true, err
}