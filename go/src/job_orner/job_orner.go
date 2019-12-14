package main

import (
	"../cafedb"
	"../values"
	"crypto/md5"
	"encoding/hex"
	"fmt"
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
	sqlCon := cafedb.NewCon()
	defer sqlCon.Close()
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

func fromJudgeThread(listenfromJudge *net.Listener, jobMap *mutexJobMap, toJobQueue *mutexJobQueue, sqlCon *cafedb.MyCon) {
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

func doFromJudgeThread(con net.Conn, jobMap *mutexJobMap, toJobQueue *mutexJobQueue, sqlCon *cafedb.MyCon) {
	//read csv result
	dataBuf := make([]byte, 1024)
	con.Read(dataBuf)
	bufStr := string(dataBuf)
	st := strings.Split(bufStr, "\n")
	bufStr = st[0]
	errStr := st[1]
	//read code session from csv
	codeSession := getSessionId(bufStr)
	fmt.Println("pass the error to front : " + errStr)
	sessionId := strings.Split(errStr, ",")[1]
	errorMes := strings.Split(errStr, ",")[2]
	sqlCon.PrepareExec("UPDATE code_sessions SET error=? WHERE id=?", errorMes, sessionId)
	con.Write([]byte("OK\n"))
	con.Close()
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
	fmt.Println("pass the result to front : " + bufStr)
	csv := strings.Split(bufStr, ",")
	result := csv[3]
	sqlCon.PrepareExec("UPDATE code_sessions SET result=? WHERE id=?", result, sessionId)
	for i := 5; i < len(csv)-1; i += 2 {
		id := generateSession()
		caseResult := csv[i]
		caseTime := csv[i+1]
		sqlCon.PrepareExec("INSERT INTO testcase_results (id, session_id, name, result, time) VALUES(?, ?, ?, ?, ?)", id, sessionId, i, caseResult, caseTime)
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
	rand.Read(b)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
