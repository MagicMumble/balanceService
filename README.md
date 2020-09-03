# balanceService
micro service for dealing with users' balance

This microservice was written for Unix-like systems on Ubuntu 16.0 using Golang and MySQL database. There are three main entities in the database: user, charge, writeOff tables. The entity user stores username and current balance. The entity charge consists from fields user(id), fromID (who the money came from), sum (amount of money charged), finalBalance (after completing transaction) and field info with some information. The table writeOff is similar to table charge. Therefore the type of relationship between tables user-charge (user-writeOff) is one to many. 

To use this app you need to install mysql server. Firstly install it with command:

    sudo apt-get install mysql-community-server
    
Then start a server and create user with specified domen and password:

    udo service mysql start
    mysql -u root -p                                                    #enter your root rassword
    create user 'user1'@'localhost' identified by 'password1';
    
Last step is creating your databse and grant all access to the user:

    create database testdb;
    grant all on testdb.* to 'user1';
    
To create tables inside new database run file `createDB.go` from `balanceService/db`:

    go build createDB.go && ./createDB
    
To be able to get currency rates from page `https://exchangeratesapi.io/` and convert balance you need to register there and get the KEY (visit page `page https://openexchangerates.org/signup`). To use mySql driver in go program install "github.com/go-sql-driver/mysql" with the command:

    go get -u "github.com/go-sql-driver/mysql"

Run the server from `balanceService`:

    go build server.go && ./server

Let's take a look at the API methods of this app. Request for adding new user:
    
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"username": "anna"}' \
     http://localhost:9000/users/add
     
 Request for charging the user bank account:
 
     curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1", "sum": "100", "info": "scolarship"}' \
     http://localhost:9000/users/charge
     
 Request for writing off some money from the user bank account:
 
     curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1", "sum": "100", "info": "debt in bank"}' \
     http://localhost:9000/users/writeOff
     
 Request for transfering some amount of money from one user to another:
 
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"fromID": "1", "toID": "2", "sum": "100", "info": "for dinner inrestaurant"}' \
     http://localhost:9000/users/transfer
     
 Request for getting one user's current balance:
 
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1"}' \
     http://localhost:9000/users/getBalance
     
 Request for getting converted user's current balance:
 
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1"}' \
     http://localhost:9000/users/getBalance?currency=USD
     
 Request for getting report with all charge transactions of one user:
 
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1"}' \
     http://localhost:9000/users/getChargeReport
     
 Request for getting report with all write off transactions of one user:
 
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1"}' \
     http://localhost:9000/users/getWriteOffReport
     
 Request for getting report with sorted by date or sum charge/write off transactions of one user:
 
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1"}' \
     http://localhost:9000/users/getWriteOffReport?sort=date
