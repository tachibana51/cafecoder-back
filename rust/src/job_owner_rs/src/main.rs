#[macro_use]
extern crate mysql;
extern crate serde;
extern crate serde_json;
use std::thread;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::net::{TcpListener, TcpStream, Shutdown};
use std::io::{Read, Write};
use std::str;
use mysql as my;

//job map<job, num>
struct MutexJobMap {
    mutexJobMap : Mutex<HashMap<String, Vec<u8>>>,
}

impl MutexJobMap {
    fn new() -> MutexJobMap {
        MutexJobMap {
            mutexJobMap : Mutex::new(HashMap::new())
        }
    }
}

//json
use serde::{Serialize, Deserialize};
#[derive(Serialize, Deserialize)]
struct Testcase {
    name : String,
    result: String,
    memoryUsed: i64,
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

fn read_data_stream(mut stream: &TcpStream) -> [u8; 1024] {
    let mut data = [0 as u8; 1024];
    match stream.read(&mut data) {
        Err(err) => println!("tcp read Error: {}", err) ,
        _ => (),
    }
    data
}

//dial tcp

//4649port
fn handle_for_api_rq(stream: TcpStream) -> Result<String, str::Utf8Error> {
    let data = read_data_stream(&stream);
    let req_csv : Vec<&str> = str::from_utf8(&data)?.split(',').collect();
    Ok("4649 ok".to_owned())
}

//5963port
fn handle_for_judge_rq(stream: TcpStream) {
    let data = read_data_stream(&stream);
}

fn main() {
    //todo config struct
    let JUDGE_HOST_PORT: &'static str = env!("JUDGE_HOST_PORT");
    let JUDGE_MAX : &'static str = env!("JUDGE_MAX");
    let mut children = vec![];
    let mut mutex_job_map = MutexJobMap::new();
    //spawn api thread
    children.push(thread::spawn(move|| {
        let listener_from_api = TcpListener::bind("0.0.0.0:4649").unwrap();
        for stream in listener_from_api.incoming() {
            thread::spawn(move|| {
                println!("{:?}", handle_for_api_rq(stream.unwrap()).unwrap_or("4649 err".to_owned()));
            });
        }
    }));
    // spawn judge thread
    children.push(thread::spawn(move|| {
        let listener_from_judge = TcpListener::bind("0.0.0.0:5963").unwrap();
        for stream in listener_from_judge.incoming() {
            thread::spawn(move|| {
                handle_for_judge_rq(stream.unwrap())
            });
        }
    }));

    for child in children {
        let _ = child.join();
    }
}

