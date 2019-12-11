package cafedb

import(
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
	"../values"
	"regexp"
	"reflect"
)

type MyCon struct {
	DB *sql.DB
	Regex *regexp.Regexp
}

func NewCon() (*MyCon) {
	db, err := sql.Open("mysql", values.MySQLDBN)
	if err != nil {
		panic(err.Error())
	}
	con := new(MyCon)
	con.DB = db
	con.Regex = regexp.MustCompile(`[^(0-9a-zA-Z\._@)]+`)
	return con
}

func (con MyCon) SafeSelect(sql string, bindData... interface{} ) (*sql.Rows, error) {
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
	fmt.Println(bindSql)
	rows, err := con.DB.Query(bindSql)
	if err != nil{
		return nil, err
	}
	return rows, err
}

func (con MyCon) PrepareExec(sql string, bindData... interface{}) (bool, error){
	stmt, err := con.DB.Prepare(sql)
	if err != nil {
		return false, err
	}
	stmt.Exec(bindData...)
	return true, err
}

//dont forget
func (con MyCon) Close(){
	con.DB.Close()
}