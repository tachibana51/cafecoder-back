package main

import (
	"../cafedb"
	"../values"
	"bytes"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
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
	AuthToken string `json:"auth_token"`
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
	AuthToken   string `json:"auth_token"`
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
	Result bool   `json:"result"`
	Token  string `json:"auth_token"`
}

//GET /api/v1/contest
type reqGetContest struct {
	ContestId string `json:"contest_id"`
}

type resGetContest struct {
	ContestName string `json:"contest_name"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	IsOpen      bool   `json:"is_open"`
}

//GET /api/v1/testcase
type reqGetTestCase struct {
	CodeSession string `json:"code_session"`
}

type resGetTestCase struct {
	Testcases []testcase `json:"testcases"`
}

//GET /api/v1/ranking
type firstAC struct {
	ProblemName string `json:"problm_ename"`
	SubmitId    string `json:"submit_id"`
	SubmitTime  string `json:"submit_ime"`
	Point       int    `json:"point"`
}

type contestResult struct {
	Rank     int       `json:"rank"`
	Username string    `json:"username"`
	Submits  []firstAC `json:"submits"`
	Point    int       `json:"point"`
}

type reqGetRanking struct {
	ContestId string `json:"contest_id"`
}

type resGetRanking struct {
	Ranking []contestResult `json:ranking`
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
	http.HandleFunc("/api/v1/contest", FuncWrapper(contestHandler, sqlCon))
	http.HandleFunc("/api/v1/auth", FuncWrapper(authHandler, sqlCon))
	http.HandleFunc("/api/v1/ranking", FuncWrapper(rankingHandler, sqlCon))
	http.ListenAndServe(":8080", nil)
}

//api/v1/result
func resultHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	case "GET":
		//template for request
		var jsonData reqGetResult
		body, _ := readData(&r)
		err := json.Unmarshal(body, &jsonData)
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
	   	   Code        string `json:code`
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
		err := json.Unmarshal(body, &jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		//read data from db
		rows, err := sqlCon.SafeSelect("SELECT users.id FROM users WHERE users.name = '%s' AND users.auth_token = '%s'", jsonData.Username, jsonData.AuthToken)
		if err != nil {
			fmt.Println(err)
			return
		}
		rows.Next()
		var userId string
		rows.Scan(&userId)
		if userId == "" {
			return
		}

		rows, err = sqlCon.SafeSelect("SELECT problems.id, problems.point , testcases.listpath FROM contests, problems, users, testcases WHERE problems.contest_id = contests.id AND testcases.id = problems.testcase_id AND contests.id = '%s' AND problems.name = '%s'", jsonData.ContestId, jsonData.Problem)
		if err != nil {
			fmt.Println(err)
			return
		}
		rows.Next()
		var (
			problemId    string
			lang         string
			point        int
			testcasePath string
		)
		lang = jsonData.Language
		rows.Scan(&problemId, &point, &testcasePath)
		sessionId := generateSession()
		//upload file
        /*
		filename := "/submits/" + userId + "_" + sessionId
		file, err := os.Create(fmt.Sprintf("./fileserver%s", filename))
		if err != nil {
			fmt.Println(err)
			return
		}
		file.Write([]byte(jsonData.Code))
		file.Close()
        */
        //uploadfiletest

		filename := "/home/akane/cafe/cafecoder-back/fileserver/submits/" + userId + "_" + sessionId
		file, err := os.Create(fmt.Sprintf("%s", filename))
		if err != nil {
			fmt.Println(err)
			return
		}
		file.Write([]byte(jsonData.Code))
		file.Close()
		sqlCon.PrepareExec("INSERT INTO code_sessions (id, problem_id, user_id, lang, result,upload_date) VALUES(?, ?, ?, ?, 'WJ', NOW())", sessionId, problemId, userId, lang)

		//con job_order
		con, err := net.Dial("tcp", values.QueHostPort)
		if err != nil {
			fmt.Println(err)
			return
		}
		argStr := []string{"dummy", sessionId, filename, lang, testcasePath, fmt.Sprint(point)}
		con.Write([]byte(strings.Join(argStr, ",")))
		con.Close()
		//convert to Json
		res := resPostCode{CodeSession: sessionId}
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

//api/v1/testcase
func testcaseHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	case "GET":
		//template for request
		var jsonData reqGetTestCase
		body, _ := readData(&r)
		err := json.Unmarshal(body, &jsonData)
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
		err := json.Unmarshal(body, &jsonData)
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
		result := (username != "")
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
		err := json.Unmarshal(body, &jsonData)
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
		rows.Next()
		rows.Scan(&userid)
		if userid != "" {
			return
		}
		userId := generateSession()
		username := jsonData.Username
		passwordHash := cafedb.GetHash(jsonData.Password)
		sqlCon.PrepareExec("INSERT INTO users (id, name, password_hash, role) VALUES (?, ?, ?, 'user')", userId, username, passwordHash)
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

//api/v1/token
func authHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	case "POST":
		//template for request
		var jsonData reqPostAuth
		body, _ := readData(&r)
		err := json.Unmarshal(body, &jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		//auth
		hash := cafedb.GetHash(jsonData.Password)
		rows, err := sqlCon.SafeSelect("SELECT users.id FROM users WHERE users.name = '%s' AND users.password_hash = '%s'", jsonData.Username, hash)
		if err != nil {
			fmt.Println(err)
			return
		}
		var res resPostAuth
		var userId string
		rows.Next()
		rows.Scan(&userId)
		res.Result = (userId != "")
		res.Token = cafedb.GetHash(generateSession())
		//set token
		if res.Result {
			sqlCon.PrepareExec("UPDATE users SET auth_token=? WHERE id=?", res.Token, userId)
		}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, string(jsonBytes))
	}

}

//GET /api/v1/contest
func contestHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	case "GET":
		var jsonData reqGetContest
		body, _ := readData(&r)
		err := json.Unmarshal(body, &jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		//get db
		rows, err := sqlCon.SafeSelect("SELECT contests.name FROM contests WHERE contests.id = '%s'", jsonData.ContestId)
		if err != nil {
			fmt.Println(err)
			return
		}
		var contestName string
		var res resGetContest
		rows.Next()
		rows.Scan(&contestName)
		res.ContestName = contestName
		rows, err = sqlCon.SafeSelect("SELECT contests.name FROM contests WHERE contests.id = '%s' AND NOW() > contests.start_time", jsonData.ContestId)
		rows.Scan(&contestName)
		res.IsOpen = (contestName != "")
		//convert to json
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, string(jsonBytes))
	}
}

func rankingHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
	switch r.Method {
	case "GET":
		var jsonData reqGetRanking
		body, _ := readData(&r)
		err := json.Unmarshal(body, &jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		//get first ACs
		// name sessionid date point
		var result []contestResult
		var acs []firstAC
		var userIds []string
		var userNames []string
		_, err = sqlCon.SafeSelect("CREATE OR REPLACE VIEW cafecoder.%s AS SELECT users.id userid, problems.name problem, (SELECT code_sessions.id FROM code_sessions WHERE code_sessions.problem_id=problems.id AND problems.name = problem) sessionid, (SELECT code_sessions.upload_date FROM code_sessions WHERE code_sessions.id=sessionid) upload_date,  problems.point point, SUM(problems.point) sumpoint, code_sessions.result FROM contests, problems, code_sessions, users WHERE contests.id = problems.id AND users.id = code_sessions.user_id AND problems.contest_id = '%s' AND code_sessions.problem_id = problems.id AND (code_sessions.upload_date BETWEEN contests.start_time AND contests.end_time) AND code_sessions.result = 'AC' GROUP BY problems.id, users.id HAVING upload_date= MIN(upload_date) ORDER BY sumpoint DESC, upload_date ASC", jsonData.ContestId, jsonData.ContestId)
		if err != nil {
			fmt.Println(err)
			return
		}

		rows, err := sqlCon.SafeSelect("SELECT userid name FROM cafecoder.%s, users WHERE userid = users.id", jsonData.ContestId)
		if err != nil {
			fmt.Println(err)
			return
		}
		var userId string
		var userName string
		for rows.Next() {
			rows.Scan(&userId, &userName)
			userIds = append(userIds, userId)
			userNames = append(userNames, userName)
		}

		for i, userid := range userIds {
			rows, err := sqlCon.SafeSelect("SELECT problem, sessionid, upload_date, point FROM cafecoder.%s WHERE cafecoder.%s.userid = '%s' AND users.id = cafecoder.%s.userid", jsonData.ContestId, userid, jsonData.ContestId)
			if err != nil {
				fmt.Println(err)
				return
			}
			j := 0
			sumOfPoint := 0
			for rows.Next() {
				acs = append(acs, *new(firstAC))
				rows.Scan(&acs[j].ProblemName, &acs[j].SubmitId, &acs[j].SubmitTime, &acs[j].Point)
				sumOfPoint += acs[j].Point
				j += 1
			}
			result = append(result, *new(contestResult))
			result[i].Rank = i
			result[i].Username = userNames[i]
			result[i].Submits = acs
			result[i].Point = sumOfPoint
		}
		//convert to json
		jsonBytes, err := json.Marshal(result)
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
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
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
