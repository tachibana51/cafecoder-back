#[macro_use]
extern crate mysql;
extern crate serde;
extern crate serde_json;
use std::thread;
use std::net::{TcpListener, TcpStream, Shutdown};
use std::sync::mpsc::channel;
use std::io::{Read, Write};
use std::str;
use mysql as my;
//json
use serde::{Serialize, Deserialize};
#[derive(Serialize, Deserialize)]
struct Testcase {
    name : String,
    result: String,
    memoryUsed: i64,
    time: i64,
}


fn read_data_stream(mut stream: &TcpStream) ->  [u8; 1024] {
    let mut data = [0 as u8; 1024];
    match stream.read(&mut data) {
        Err(err) => println!("tcp read Error: {}", err) ,
        _ => (),
    }
    return data;
}

fn handle_for_api_rq(stream: TcpStream) {
    let data = read_data_stream(&stream);
    let req_csv : Vec<&str> = str::from_utf8(&data).unwrap().split(',').collect(); 
}

fn handle_for_judge_rq(stream: TcpStream) {
    let data = read_data_stream(&stream);
}

fn main() {
    let mut children = vec![];
    //spawn api thread
    children.push(thread::spawn(move|| {
        let listener_from_api = TcpListener::bind("0.0.0.0:4649").unwrap();
        for stream in listener_from_api.incoming() {
            thread::spawn(move|| {
                handle_for_api_rq(stream.unwrap())
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

