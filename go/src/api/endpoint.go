package main

import (
	"../cafedb"
	"../values"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
)

type testcase struct {
	CaseName string `json:"testcase_name"`
	Result   string `json:"result"`
	Runtime  int    `json:"runtime"`
}

//POST /api/v1/code
type reqPostCode struct {
	Code      string `json:"code"`
	Username  string `json:"username"`
	Problem   string `json:"problem"`
	Language  string `json:"language"`
	ContestId string `json:"contest_id"`
}

type resPostCode struct {
	CodeSession string `json:"code_session"`
}

//GET /api/v1/result
type reqGetResult struct {
	CodeSession string `json:"code_session"`
}

type resGetResult struct {
	Username    string `json:"username"`
	ContestName string `json:"contestname"`
	Problem     string `json:"problem"`
	Point       string `json:"point"`
	Lang        string `json:"language"`
	Result      string `json:"result"`
}

//GET /api/v1/user
type reqGetUser struct {
	Username string `json:"username"`
}

type resGetUser struct {
	Result bool `json:"result"`
}

//POST /api/v1/user
type reqPostUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type resPostUser struct {
	Result bool `json:"result"`
}

//POST /api/v1/auth
type reqPostAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type resPostAuth struct {
	Result bool `json:"result"`
}

//GET /api/v1/contest
type reqGetContest struct {
	ContestId string `json:"contest_id"`
}

type resGetContest struct {
	ContestName string `json:"contest_name"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	isOpen      bool   `json:"is_open"`
}

//GET /api/v1/testcase
type reqGetTestCase struct {
	CodeSession string `json:"code_session"`
}

type resGetTestCase struct {
	Testcases []testcase `json:"testcases"`
}

type TrashScanner struct{}

func (TrashScanner) Scan(interface{}) error {
	return nil
}

func main() {
	evenvListenerThread()
}

func evenvListenerThread() {
	sqlCon := cafedb.NewCon()
	defer sqlCon.Close()
	http.HandleFunc("/api/v1/result", FuncWrapper(resultHandler, sqlCon))
	http.HandleFunc("/api/v1/code", FuncWrapper(codeHandler, sqlCon))
	http.HandleFunc("/api/v1/testcase", FuncWrapper(testcaseHandler, sqlCon))
	http.HandleFunc("/api/v1/user", FuncWrapper(userHandler, sqlCon))
	//http.HandleFunc("/api/v1/contest", FuncWrapper(contestHandler, sqlCon))
	http.HandleFunc("/api/v1/auth", FuncWrapper(authHandler, sqlCon))
	http.ListenAndServe(":8080", nil)
}

//api/v1/result
func resultHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	case "GET":
		//template for request
		var jsonData reqGetResult
		body, _ := readData(&r)
		err := json.Unmarshal(body, jsonData)
		//read data from db
		rows, err := sqlCon.SafeSelect("SELECT users.name, contests.name, problems.name, problems.point, code_sessions.lang, code_sessions.result  FROM contests, problems, users WHERE sessions.id = '%s' AND problems.contest_id = contests.id  AND code_sessions.user_id = users.id ", jsonData.CodeSession)
		if err != nil {
			fmt.Println(err)
			return
		}
		var res resGetResult
		rows.Next()
		rows.Scan(&res.Username, &res.ContestName, &res.Problem, &res.Point, &res.Lang, &res.Result)
		//convert to Json
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, string(jsonBytes))

	default:
		return
	}

}

//api/v1/code
func codeHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	/*
	   in
	       Username    string `json:"username"`
	       Problem     string `json:"problem"`
	       Language    string `json:"language"`
	       ContestId   string `json:"contest_id"`

	   out
	       CodeSession string `json:"code_session"`
	*/
	case "POST":
		//template for request
		var jsonData reqPostCode
		body, _ := readData(&r)
		err := json.Unmarshal(body, jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		//read data from db
		rows, err := sqlCon.SafeSelect("SELECT problems.id, users.id, code_sessions.lang , testcases.listpath FROM contests, problems, users, testases WHERE problems.contest_id = contests.id AND testcases.id = problems.testcase_id AND contests.id = '%s' AND problems.name = '%s'", jsonData.ContestId, jsonData.Problem)
		rows.Next()
		var (
			problemId    string
			userId       string
			lang         string
			testcasePath string
		)
		rows.Scan(&problemId, &userId, &lang, &testcasePath)
		sessionId := generateSession()
		//upload file
		filename := "/submits/" + userId + "-" + sessionId
		file, err := os.Create(fmt.Sprintf("./fileserver%s", filename))
		if err != nil {
			fmt.Println(err)
			return
		}
		file.Write([]byte(jsonData.Code))
		file.Close()
		sqlCon.PrepareExec("INSERT INTO code_sessions (id, problem_id, user_id, lang, upload_date) VALUES(?, ?, ?, ?, NOW())", sessionId, problemId, userId, lang)

		//con job_order
		con, err := net.Dial("tcp", values.QueHostPort)
		if err != nil {
			fmt.Println(err)
			return
		}
		argStr := []string{"dummy", sessionId, filename, lang, testcasePath, "point"}
		con.Write([]byte(strings.Join(argStr, ",")))
		con.Close()
		//convert to Json
		jsonBytes, err := json.Marshal(sessionId)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, string(jsonBytes))

	default:
		return
	}
}

//api/v1/testcase
func testcaseHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	case "GET":
		//template for request
		var jsonData reqGetTestCase
		body, _ := readData(&r)
		err := json.Unmarshal(body, jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		//read results from db
		rows, err := sqlCon.SafeSelect("SELECT testcase_results.name, testcase_results.result, testcase_results.time FROM testcase_rsults WHERE testcases_result.code_session='%s' AND test", jsonData.CodeSession)
		if err != nil {
			fmt.Println(err)
			return
		}
		caseList := make([]testcase, 100)
		i := 0
		for rows.Next() {
			if i >= 100 {
				return
			}
			var t testcase
			rows.Scan(&t.CaseName, &t.Result, &t.Runtime)
			caseList[i] = t
			i += 1
		}
		//convert to Json
		jsonBytes, err := json.Marshal(caseList[:i])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, string(jsonBytes))
	}
}

//api/v1/user
func userHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	case "GET":
		//template for request
		var jsonData reqGetUser
		body, _ := readData(&r)
		err := json.Unmarshal(body, jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		//read result from db
		rows, err := sqlCon.SafeSelect("SELECT users.id FROM users WHERE users.name = '%s'", jsonData.Username)
		if err != nil {
			fmt.Println(err)
			return
		}
		rows.Next()
		var username string
		rows.Scan(&username)
		result := (username == "")
		res := resGetUser{Result: result}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, string(jsonBytes))

	case "POST":
		//template for request
		var jsonData reqPostUser
		body, _ := readData(&r)
		err := json.Unmarshal(body, jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		//is exists
		rows, err := sqlCon.SafeSelect("SELECT users.id FROM users WHERE users.name = '%s'", jsonData.Username)
		if err != nil {
			fmt.Println(err)
			return
		}
		var userid string
		rows.Scan(&userid)
		if userid != "" {
			return
		}
		userId := generateSession()
		username := jsonData.Username
		passwordHash := cafedb.GetHash(jsonData.Password)
		sqlCon.PrepareExec("INSERT INTO users (id, name, password_hash, role) VALUES (?, ?, ?, user)", userId, username, passwordHash)
		//conver to json
		res := resGetUser{Result: true}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, string(jsonBytes))
	}
}

func authHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	case "POST":
		//template for request
		var jsonData reqPostAuth
		body, _ := readData(&r)
		err := json.Unmarshal(body, jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		hash := cafedb.GetHash(jsonData.Password)
		rows, err := sqlCon.SafeSelect("SELECT users.id FROM users WHERE users.name = '%s' AND users.password_hash = '%s'", jsonData.Username, hash)
		if err != nil {
			fmt.Println(err)
			return
		}
		var res resPostAuth
		var userid string
		rows.Scan(&userid)
		res.Result = (userid != "")
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, string(jsonBytes))
	}

}
func FuncWrapper(f interface{}, c *cafedb.MyCon) func(http.ResponseWriter, *http.Request) {
	function := f.(func(http.ResponseWriter, *http.Request, *cafedb.MyCon))
	return func(w http.ResponseWriter, r *http.Request) { function(w, r, c) }
}

func generateSession() string {
	b := make([]byte, 8)
	h := md5.New()
	rand.Read(b)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

//todo buffer reader
func readData(r **http.Request) ([]byte, error) {
	body := make([]byte, 1000000)
	body, err := ioutil.ReadAll((*r).Body)
	if err != nil {
		fmt.Println(err)
		return body, err
	}
	//read json
	body = bytes.Trim(body, "\x00")
	return body, err
}
