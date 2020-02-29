#[macro_use]
extern crate mysql;
extern crate serde;
extern crate serde_json;
use std::collections::VecDeque;
use std::thread;
use std::collections::HashMap;
use std::sync::{Arc, Mutex, RwLock};
use std::net::{TcpListener, TcpStream};
use std::io::{Read, Write};
use std::str;
use mysql as my;
use serde::{Serialize, Deserialize};

//todo config struct
static JUDGE_HOST_PORT: &'static str = env!("JUDGE_HOST_PORT");
//const JUDGE_MAX : i64 = env!("JUDGE_MAX").parse().unwrap();
const JUDGE_MAX : i64 = 10;

//json
#[derive(Serialize, Deserialize)]
struct Testcase {
    name : String,
    result: String,
    memory_used: i64,
    time: i64,
}

#[derive(Serialize, Deserialize)]
struct OverAllResult {
    session_id : String,
    over_all_time : i64,
    over_all_result : String,
    over_all_score : i64,
    err_message : String,
    testcases : Vec<Testcase>,
}

fn read_data_stream(mut stream: TcpStream) -> [u8; 1024] {
    let mut data = [0 as u8; 1024];
    match stream.read(&mut data) {
        Err(err) => println!("tcp read Error: {}", err) ,
        _ => (),
    }
    data
}

//dial tcp
fn pass_to_judge (data : &[u8]) -> Result<usize, std::io::Error> {
    TcpStream::connect(JUDGE_HOST_PORT)?.write(data)
}

//4649port
//todo make a argument trait
fn handle_for_api_rq(stream: TcpStream, r_m_jobmap : *const RwLock<HashMap<String,i64>>, r_m_jobqueue : *const RwLock<VecDeque<[u8; 1024]>>) -> Result<String, String> {
    let data = read_data_stream(stream);
    let req_csv : Vec<&str> = str::from_utf8(&data).map_err(|e| e.to_string())?.split(',').collect();
    //deside judge or que
    if req_csv.len() <= 1 {
        return Err("csv parse error".to_owned());
    }
    let mut current_map = (*r_m_jobmap).write().unwrap();
    if let (i @ 0...JUDGE_MAX) = current_map.keys().len() as i64 {
        return match pass_to_judge(&data) {
            //todo remofe from map then pass judge err 
            Ok(_x) => {
                current_map.insert(req_csv[1].to_string(), i);
                println!("now works {}", i);
                Ok("4649 ok".to_owned())
               }
            Err(_err) => Err("pass judge write Error".to_owned()),
        }
    }else{
        //push to queue
        let mut current_queue = (*r_m_jobqueue).write().unwrap();
        println!("to queue {:?}", current_map);
        current_queue.push_front(data);
        return Ok("to queue".to_owned())
    }
}

//5963port
fn handle_for_judge_rq(stream: TcpStream, r_m_jobmap : *const RwLock<HashMap<String,i64>>, r_m_jobqueue : *const RwLock<VecDeque<[u8; 1024]>>) -> Result<String, String> {
    let data = read_data_stream(stream);
    Ok("test".to_owned())
}

fn main() {
    let mut children = vec![];
    let m_jobmap  = Arc::new(RwLock::new(HashMap::<String, i64>::new()));
    let m_jobqueue = Arc::new(RwLock::new(VecDeque::<[u8; 1024]>::new()));
    //spawn api thread
    children.push(thread::spawn(move|| {
        let listener_from_api = TcpListener::bind("0.0.0.0:4649").unwrap();
        for stream in listener_from_api.incoming() {
                let r_m_jobmap : *const RwLock<HashMap<String,i64>> = Arc::into_raw(m_jobmap);
                let r_m_jobqueue : *const RwLock<VecDeque<[u8; 1024]>> = Arc::into_raw(m_jobqueue);
                unsafe {thread::spawn(move|| {
                    println!("{:?}", handle_for_api_rq(stream.unwrap(), r_m_jobmap, r_m_jobqueue).unwrap_or_else(|e| e.to_owned()));
                });}
        }
    }));
    // spawn judge thread
    children.push(thread::spawn(move|| {
            let listener_from_judge = TcpListener::bind("0.0.0.0:5963").unwrap();
            for stream in listener_from_judge.incoming() {
            thread::spawn(move|| {
            let r2_m_jobmap : *const RwLock<HashMap<String,i64>> = Arc::into_raw(m_jobmap);
            let r2_m_jobqueue : *const RwLock<VecDeque<[u8; 1024]>> = Arc::into_raw(m_jobqueue);
                println!("{:?}", handle_for_judge_rq(stream.unwrap(), r2_m_jobmap, r2_m_jobqueue).unwrap_or_else(|e| e.to_owned()));
            });
        }
    }));

    for child in children {
        let _ = child.join();
    }
}

