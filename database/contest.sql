CREATE TABLE cafecoder.users( 
    id varchar(32) NOT NULL,
    name varchar(100) NOT NULL,
    email varchar(255),
    password_hash varchar(64) NOT NULL,
    role varchar(10) NOT NULL,
    rate int,
    PRIMARY KEY (id)
);

CREATE TABLE cafecoder.contests(
    id varchar(32) NOT NULL,
    name varchar(32) NOT NULL,
    start_time datetime NOT NULL,
    end_time datetime NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE cafecoder.problems(
    id varchar(32) NOT NULL,
    contest_id varchar(4) NOT NULL,
    name varchar(4) NOT NULL,
    point int,
    testcase_id varchar(32),
    PRIMARY KEY (id)
);

CREATE TABLE cafecoder.code_sessions(
    id varchar(32) NOT NULL,
    problem_id varchar(32) NOT NULL,
    user_id varchar(32) NOT NULL,
    lang varchar(32) NOT NULL,
    upload_date datetime,
    result varchar(8), 
    PRIMARY KEY (id)
);

CREATE TABLE cafecoder.testcases(
    id varchar(32) NOT NULL, 
    problem_id varchar(32) NOT NULL,
    listpath varchar(1024) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE cafecoder.testcase_results(
    id varchar(32) NOT NULL,
    session_id(32) NOT NULL,
    name varchar(255) NOT NULL,
    result varchar(8),
    time int,
    PRIMARY KEY (id)
)
