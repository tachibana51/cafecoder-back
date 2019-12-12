package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "../values"
    "net"
    "io/ioutil"
    "../cafedb"
    "strings"
    "bytes"
)
type postCode struct {
    Code        string `json:"code"`
    CodeSession string `json:"code_session"`
    UserId      string `json:"user_id"`
    Language    string `json:"language"`
    ContestId   string `json:"contest_id"`
    Problem     string `json:"problem"`
}

type testcase struct {
    CaseName    string `json:"testcase_name"`
    Result      string `json:"result"`
    Runtime     int        `json:"runtime"`
}

type getResult struct {
    CodeSession string `json:"code_session"`
    UserId      string `json:"user_id"`
}

type resGetResult struct {
    CodeSession string `json:"code_session"`
    UserId      string `json:"user_id"`
    ContestId   string `json:"contestId"`
    Problem     string `json:"problem"`
    Code        string `json:"code"`
    Lang        string `json:"language"`
    Result      string `json:"result"`    
    Testcases   []testcase `json:"testcases"`
}

type TrashScanner struct{}

func (TrashScanner) Scan(interface{}) error {
    return nil
}

func resultHandler(w http.ResponseWriter, r *http.Request,sqlCon *cafedb.MyCon) {
    switch (r.Method) {
        case "GET":
            body := make([]byte, 4096)
            body, err := ioutil.ReadAll(r.Body)
            defer r.Body.Close()
            if err != nil {
                fmt.Println(err)
                return
            }
            //read json
            body = bytes.Trim(body,"\x00")
            var jsonData getResult
            err = json.Unmarshal(body, &jsonData)
            if err != nil {
                fmt.Println(err)
                return
            }
            //read data from db
            rows, err := sqlCon.SafeSelect("SELECT code_session, user_id, contest_id, problem, lang, result  FROM uploads WHERE code_session='%s' AND user_id='%s'", jsonData.CodeSession, jsonData.UserId)
            if err != nil {
                fmt.Println(err)
                return
            }
            var res resGetResult
            rows.Next()
            rows.Scan(&res.CodeSession, &res.UserId, &res.ContestId, &res.Problem, &res.Lang, &res.Result)
            formatBytes, err := json.Marshal(res)
            if err != nil {
                fmt.Println(err)
                return
            }
            fmt.Fprintf(w, string(formatBytes))

        default:
            return
    }

}

//not work yet
func codeHandler(w http.ResponseWriter, r *http.Request, sqlCon *cafedb.MyCon) {
    switch (r.Method) {
        case "POST":
            body := make([]byte, 4096)
            jsonData := postCode{}
            _, err := r.Body.Read(body)
            //read json
            err = json.Unmarshal(body, &jsonData)
            if err != nil {
                fmt.Println(err)
                return
            }
            //read data from db
            sqlCon.SafeSelect("SELECT testcase_list_dir FROM problem WHERE problem_id='%s'", jsonData.Problem)

            //con job_order
            con, err := net.Dial("tcp", values.QueHostPort)
            if err != nil {
                fmt.Println(err)
                return
            }
            argStr := []string{jsonData.Code, jsonData.CodeSession, "path", jsonData.Language, "testcasedir", "point"}
            con.Write([]byte(strings.Join(argStr ,",")))
            con.Close()
            fmt.Fprintf(w, "OK")

        default:
            return
    }
} 

func FuncWrapper(f interface{}, c *cafedb.MyCon) func(http.ResponseWriter, *http.Request) {
    function := f.(func(http.ResponseWriter, *http.Request, *cafedb.MyCon))
    return  func(w http.ResponseWriter, r *http.Request) { function(w, r, c) }
}

func evenvListenerThread() {
    sqlCon := cafedb.NewCon()
    defer sqlCon.Close()
    http.HandleFunc("/api/v1/result", FuncWrapper(resultHandler, sqlCon))
    http.HandleFunc("/api/v1/code", FuncWrapper(codeHandler, sqlCon))
    http.ListenAndServe(":8080", nil)
}

func main()  {
    evenvListenerThread()    
}