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
	FRONT_HOST_PORT = "localhost:80"
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
		fmt.Println("bind err front thread")
	}

	listenfromJudge, err := net.Listen("tcp", "0.0.0.0:5963")
	if err != nil {
		fmt.Println("bind err judge thread")
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
	codeSession := getSessionId(bufStr)
	//block race condition
	jobMap.Lock()
	now_worker := len(jobMap.dictionary) 
	if now_worker < MAX_WORKER {
		//pass the job
		fmt.Println("pass the job judge : " + codeSession)
		con.Write([]byte("JUDGE\n"))
		go passJobToJudge(string(dataBuf))
		jobMap.dictionary[codeSession] = now_worker 
	}else{
		toJobQueue.Lock()
		//add que
		fmt.Println("add the job to que : " + codeSession)
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
			fmt.Println(err)
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
	codeSession := getSessionId(bufStr)
	if codeSession == "error" {
		fmt.Println("pass the error to front : " + bufStr)
		go passResultToFront(bufStr)
		con.Write([]byte("OK\n"))
		con.Close()
		return 
	}
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
	go passResultToFront(bufStr)
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

func passResultToFront(arg string){
	conn , err := net.Dial("tcp", FRONT_HOST_PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Write([]byte(arg + "\n"))
	conn.Close()
}
func getSessionId(str string) (string) {
	return strings.Split(str, ",")[0]
}