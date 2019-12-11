CREATE TABLE cafecoder.contests(
  contest_id varchar(32) NOT NULL,
  contest_name varchar(32) NOT NULL,
  start_time datetime NOT NULL,
  end_time datetime NOT NULL,
 PRIMARY KEY (contest_id)
);
CREATE TABLE cafecoder.problem(
  contest_id varchar(32) NOT NULL,
  problem_id varchar(4) NOT NULL,
  point int,
  testcase_list_dir varchar(255),
  PRIMARY KEY (contest_id, problem_id)
);
CREATE TABLE cafecoder.uploads (
    code_session varchar(32) NOT NULL,
    contest_id varchar(32) NOT NULL,
    problem varchar(1) NOT NULL,
    user_id varchar(32) NOT NULL,
    lang varchar(32) NOT NULL,
    upload_date datetime,
    result varchar(8),
    PRIMARY KEY (code_session)
);
CREATE TABLE cafecoder.users 
	(uid varchar(32) NOT NULL,
	 username varchar(100) NOT NULL,
	 email varchar(255),
	 password_hash varchar(64) NOT NULL,
	 rate int,
     role varchar(10) NOT NULL,
	 PRIMARY KEY (uid)
);
