#maniawase todo contestid
CREATE TABLE kakecoder.tea002score(
  uid varchar(32),
  score int,
  PRIMARY KEY (uid), 
);
CREATE TABLE kakecoder.tea002uploads (
    uid varchar(32),
    problem varchar(1),
    code_session varchar(255),
    user_id varchar(32),
    PRIMARY KEY (uid)
);
CREATE TABLE kakecoder.tea002rank (
  uid varchar(32),
  score int,
  sumtime int,
  PRIMARY KEY (uid) 
);
