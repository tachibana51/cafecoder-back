package main

import (
	"../cafedb"
    "encoding/json"
	"../values"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"net"
	"strings"
	"sync"
)

const (
	MaxWorker     = values.MaxWorker
	JudgeHostPort = values.JudgeHostPort
	FrontHostPort = values.FrontHostPort
)

type overAllResultJSON struct {
	SessionID     string            `json:"sessionID"`
	OverAllTime   int64             `json:"time"`
	OverAllResult string            `json:"result"`
	OverAllScore  int               `json:"score"`
	ErrMessage string `json:"errMessage"`
	Testcases      []testcaseJSON `json:"testcases"`
}

type testcaseJSON struct {
	Name       string `json:"name"`
	Result     string `json:"result"`
	MemoryUsed int64  `json:"memory_used"`
	Time       int64  `json:"time"`
}



type mutexJobMap struct {
	sync.Mutex
	dictionary map[string]int
}

type mutexJobQueue struct {
	sync.Mutex
	que []string
}

func newMutexJobMap() *mutexJobMap {
	return &mutexJobMap{
		dictionary: make(map[string]int),
	}
}

func newMutexJobQueue() *mutexJobQueue {
	return &mutexJobQueue{
		que: make([]string, 0),
	}

}

func main() {
	wg := sync.WaitGroup{}
	jobMap := newMutexJobMap()
	toJobQueue := newMutexJobQueue()
	sq := cafedb.NewCon()
    sqlCon := &sq
	defer sq.Close()
	listenfromFront, err := net.Listen("tcp", "0.0.0.0:4649")
	if err != nil {
		fmt.Println("bind err front thread")
	}

	listenfromJudge, err := net.Listen("tcp", "0.0.0.0:5963")
	if err != nil {
		fmt.Println("bind err judge thread")
	}

	wg.Add(1)
	go fromFrontThread(&listenfromFront, jobMap, toJobQueue)
	wg.Add(1)
	go fromJudgeThread(&listenfromJudge, jobMap, toJobQueue, sqlCon)
	wg.Wait()
}

//todo reset queue
func initFromDB(toJobQueue *mutexJobQueue,  sqlCon **cafedb.MyCon){
    rows, err := (*sqlCon).SafeSelect("SELECT code_sessions.id FROM code_sessions WHERE code_sessions.result='WJ'")
    if err != nil {
        fmt.Println(err)
    }
    var bufStr string
    for rows.Next() {
        rows.Scan(&bufStr)
		toJobQueue.que = append(toJobQueue.que, bufStr)
    }
}
func fromFrontThread(listenfromFront *net.Listener, jobMap *mutexJobMap, toJobQueue *mutexJobQueue) {
	for {
		con, err := (net.Listener)(*(listenfromFront)).Accept()
		fmt.Println("accept Front Thread")
		if err != nil {
			continue
		}
		go doFrontThread(con, jobMap, toJobQueue)

	}
}

func doFrontThread(con net.Conn, jobMap *mutexJobMap, toJobQueue *mutexJobQueue) {
	//read csv arguments
	dataBuf := make([]byte, 1024)
	con.Read(dataBuf)
	bufStr := string(dataBuf)
	//read code session from csv
	codeSession := strings.Split(bufStr, ",")[1]
	//block race condition
	jobMap.Lock()
	now_worker := len(jobMap.dictionary)
	if now_worker < MaxWorker {
		//pass the job
		fmt.Println("pass the job judge : " + bufStr)
		con.Write([]byte("JUDGE\n"))
		go passJobToJudge(bufStr)
		jobMap.dictionary[codeSession] = now_worker + 1
	} else {
		toJobQueue.Lock()
		//add que
		fmt.Println("add the job to que : " + bufStr)
		toJobQueue.que = append(toJobQueue.que, bufStr)
		fmt.Println(toJobQueue.que)
		toJobQueue.Unlock()
		con.Write([]byte("QUEUE\n"))
	}
	jobMap.Unlock()
	con.Close()
}

func fromJudgeThread(listenfromJudge *net.Listener, jobMap *mutexJobMap, toJobQueue *mutexJobQueue, sqlCon **cafedb.MyCon) {
	for {
		con, err := (net.Listener)(*(listenfromJudge)).Accept()
		fmt.Println("accept Judge Thread")
		if err != nil {
			fmt.Println(err)
			continue
		}
		go doFromJudgeThread(con, jobMap, toJobQueue, sqlCon)

	}
}

func doFromJudgeThread(con net.Conn, jobMap *mutexJobMap, toJobQueue *mutexJobQueue, sqlCon **cafedb.MyCon) {
	//read csv result
    var jsonResult overAllResultJSON
    json.NewDecoder(con).Decode(&jsonResult)
	con.Write([]byte("OK\n"))
	con.Close()
	//read code session from csv
	codeSession := jsonResult.SessionID 
	//block race condition
	jobMap.Lock()
	//remove from jobMap
	delete(jobMap.dictionary, codeSession)
	jobMap.Unlock()
	fmt.Println(toJobQueue.que)
	if len(toJobQueue.que) != 0 {
		//dequeue
		toJobQueue.Lock()
		job := toJobQueue.que[0]
		toJobQueue.que = toJobQueue.que[1:]
		fmt.Println("pass the job to judge : " + job)
		toJobQueue.Unlock()
		//pass to judge
		go passJobToJudge(job)
	}
	result := jsonResult.OverAllResult
	(*sqlCon).PrepareExec("UPDATE code_sessions SET result=? , error=? WHERE code_sessions.id=?", result,jsonResult.ErrMessage, codeSession)
    for _, testcase := range jsonResult.Testcases {
		id := generateSession()
		caseResult := testcase.Result
        caseTime := testcase.Time
        caseName := testcase.Name
		(*sqlCon).PrepareExec("INSERT INTO testcase_results (id, session_id, name, result, time) VALUES(?, ?, ?, ?, ?)", id, codeSession, caseName, caseResult, caseTime)
	}
	con.Write([]byte("OK\n"))
	con.Close()
}

func passJobToJudge(arg string) {
	conn, err := net.Dial("tcp", JudgeHostPort)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Write([]byte(arg + "\n"))
	conn.Close()
}

func passResultToFront(arg string) {
	conn, err := net.Dial("tcp", FrontHostPort)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Write([]byte(arg + "\n"))
	conn.Close()
}

func getSessionId(str string) string {
	return strings.Split(str, ",")[0]
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
