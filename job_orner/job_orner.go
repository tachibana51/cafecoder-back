package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	MAX_WORKER = 30
	JUDGE_HOST_PORT = "localhost:8888"
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

func newMutexJobQueue() *mutexJobQueue{
	return &mutexJobQueue{
        que: make([]string, 0),
    }

}

func main() {
	wg := sync.WaitGroup{}
	jobMap := newMutexJobMap()
	toJobQueue := newMutexJobQueue()
	listenfromFront, err := net.Listen("tcp", "0.0.0.0:4649")
	if err != nil {
		fmt.Println("bind err")
	}
	listenfromJudge, err := net.Listen("tcp", "0.0.0.0:5963")
	if err != nil {
		fmt.Println("bind err")
	}
	wg.Add(1)
	go fromFrontThread(&listenfromFront, jobMap, toJobQueue)
	wg.Add(1)
	go fromJudgeThread(&listenfromJudge, jobMap, toJobQueue)
	wg.Wait()
}

func fromFrontThread(listenfromFront *net.Listener, jobMap *mutexJobMap, toJobQueue *mutexJobQueue) {
	for {
		con, err := (net.Listener)(*(listenfromFront)).Accept()
		fmt.Println("accept Front Thread");
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
	code_session := getsessionId(bufStr)
	//block race condition
	jobMap.Lock()
	now_worker := len(jobMap.dictionary) 
	if now_worker < MAX_WORKER {
		//pass the job
		fmt.Println("pass the job judge : " + code_session)
		con.Write([]byte("JUDGE\n"))
		passJobToJudge(string(dataBuf))
		jobMap.dictionary[code_session] = now_worker 
	}else{
		toJobQueue.Lock()
		//add que
		fmt.Println("add the job to que : " + code_session)
		toJobQueue.que = append(toJobQueue.que, bufStr)
		fmt.Println(toJobQueue.que)
		toJobQueue.Unlock()
		con.Write([]byte("QUEUE\n"))
	}
	jobMap.Unlock()
	con.Close()
}

func fromJudgeThread(listenfromJudge *net.Listener, jobMap *mutexJobMap, toJobQueue *mutexJobQueue){
	for {
		con, err := (net.Listener)(*(listenfromJudge)).Accept()
		fmt.Println("accept Judge Thread");
		if err != nil {
			continue
		}
		go doFromJudgeThread(con, jobMap, toJobQueue)

	}
}

func doFromJudgeThread(con net.Conn, jobMap *mutexJobMap, toJobQueue *mutexJobQueue){
	//read csv result
	dataBuf := make([]byte, 1024)
	con.Read(dataBuf)
	bufStr := string(dataBuf)
	//read code session from csv
	code_session := getsessionId(bufStr)
	//block race condition	
	jobMap.Lock()
	//remove from jobMap
	delete(jobMap.dictionary, code_session)
	jobMap.Unlock()
	fmt.Println(toJobQueue.que)
	if len(toJobQueue.que) != 0 {
		//dequeue
		toJobQueue.Lock()
		job := toJobQueue.que[0]
		toJobQueue.que = toJobQueue.que[1:]
		fmt.Println("pass the job judge : " + job)
		toJobQueue.Unlock()
		passJobToJudge(job)
	}
	con.Write([]byte("OK\n"))
	con.Close()
}

func passJobToJudge(arg string){
	conn , err := net.Dial("tcp", JUDGE_HOST_PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Write([]byte(arg + "\n"))
	conn.Close()
}

func getsessionId(str string) (string) {
	return strings.Split(str, ",")[0]
}