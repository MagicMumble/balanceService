use testdb;

create table user (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username varchar(30) unique,
  balance decimal(10,2) default 0
);

create table charge (
  user int not null,
  fromID int default -1,
  sum decimal(10,2) CHECK(sum>0),
  info varchar(100) default "no info",
  finalBalance decimal(10,2),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

create table writeOff (user int not null,
  toID int default -1,
  sum decimal(10,2) CHECK(sum>0),
  info varchar(100) default "no info",
  finalBalance decimal(10,2) check(finalBalance>=0),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
