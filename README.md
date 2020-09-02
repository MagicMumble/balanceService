# balanceService
micro service for dealing with users' balance

This microservice was written for Unix-like systems on Ubuntu 16.0 using Golang and MySQL database. There are several main entities:

user, cahrge, writeOff ........

Request for adding new user:
    
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"username": "anna"}' \
     http://localhost:9000/users/add
     
 Request for charging the user bank account:
 
     curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1", "sum": "100"}' \
     http://localhost:9000/users/charge
     
 Request for writing off some money from the user bank account:
 
     curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1", "sum": "100"}' \
     http://localhost:9000/users/writeOff
     
 Request for transfering some amount of money from one user to another:
 
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"fromID": "1", "toID": "2", "sum": "100"}' \
     http://localhost:9000/users/tranfer
     
 Request for getting one user's current balance:
 
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1"}' \
     http://localhost:9000/users/getBalance
     
 Request for getting converted user's current balance:
 
    curl --header "Content-Type: application/json" \
     --request POST \
     --data '{"userID": "1"}' \
     http://localhost:9000/users/getConvertedBalance?currency=USD
     
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
