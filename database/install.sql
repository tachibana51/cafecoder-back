CREATE DATABASE kakecoder;
CREATE TABLE kakecoder.users 
	(uid varchar(32) NOT NULL,
	 username varchar(100) NOT NULL,
	 email varchar(255),
	 password_hash varchar(64) NOT NULL,
	 rate int,
     role varchar(10) NOT NULL,
	 PRIMARY KEY (uid)
);
CREATE TABLE kakecoder.contests
	(contestid varchar(32) NOT NULL,
	contestname varchar(32) NOT NULL,
	indexpath varchar(255),
	startdate datetime,
	enddate datetime,
	PRIMARY KEY (contestid)
);
