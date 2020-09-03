package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func Execute(sqlStmt string, db *sql.DB, message string) {
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err, sqlStmt)
		return
	} else {
		log.Println(message)
	}
}

func deleteDbs(db *sql.DB) {
	sqlStmt := `drop table user;`
	Execute(sqlStmt, db, "db deleted")
	sqlStmt = `drop table charge;`
	Execute(sqlStmt, db, "db deleted")
	sqlStmt = `drop table writeOff;`
	Execute(sqlStmt, db, "db deleted")
}

func createDbs(db *sql.DB) {
	sqlStmt := `create table user (id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, username varchar(30) unique, balance decimal(10,2) default 0);`
	Execute(sqlStmt, db, "db created")
	sqlStmt = `create table charge (user int not null, fromID int default -1, sum decimal(10,2) CHECK(sum>0), info varchar(100) default "no info", finalBalance decimal(10,2), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);` //в finalBalance никогда не будет отрицательного числа
	Execute(sqlStmt, db, "db created")
	sqlStmt = `create table writeOff (user int not null, toID int default -1, sum decimal(10,2) CHECK(sum>0), info varchar(100) default "no info", finalBalance decimal(10,2) check(finalBalance>=0), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`
	Execute(sqlStmt, db, "db created")
}

func main() {
	db, err := sql.Open("mysql", "user1:password1@/testdb")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()
	deleteDbs(db)
	createDbs(db)
}

//sudo apt-get install mysql-community-server
//sudo service mysql start
//mysql -u root -p (entering pass)
// create user 'user1'@'localhost' identified by 'password1';
//create database testdb;
// grant all on testdb.* to 'user1';


