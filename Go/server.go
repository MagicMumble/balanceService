package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

const (
	APP_ID = "f1a4327b05d743d1aaf9e84c0f757e28"              // use your own key. To get a key visit page https://openexchangerates.org/signup
)

type User struct {
	Username string `json:"username"`
}

type TransferRequest struct {
	UserID string `json:"userID"`
	Sum string `json:"sum"`
	Info string `json:"info"`
}

type TransferFromToRequest struct {
	FromID string `json:"fromID"`
	ToID string `json:"toID"`
	Sum    string `json:"sum"`
	Info string `json:"info"`
}

type UserID struct {
	UserID string `json:"userID"`
}

type Response struct {
	Response string `json:"Response"`
}

func sendResponse(w http.ResponseWriter, text string) {

	w.Header().Set("Content-Type", "application/json")
	data := Response {
		Response: text,
	}
	json.NewEncoder(w).Encode(data)
}

func checkHeader(w http.ResponseWriter, req *http.Request) *json.Decoder {

	contType := req.Header.Get("Content-Type")
	if contType != "" && contType != "application/json" {
		http.Error(w, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
		return nil
	}
	req.Body = http.MaxBytesReader(w, req.Body, 1048576)
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	return dec
}

func checkErr(err error, w http.ResponseWriter) bool {

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return true
	} else {
		return false
	}
}

func getRate(currency string, balance float64, w http.ResponseWriter) float64{

	response, err := http.Get("http://openexchangerates.org/api/latest.json?app_id="+APP_ID+"&symbols=RUB,"+currency) //would just change base but this option isn't for free :(
	checkErr(err, w)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	checkErr(err, w)

	parseData := make(map[string]interface{})
	err = json.Unmarshal([]byte(string(body)), &parseData)

	if !checkErr(err, w) {
		for left, _ := range parseData {
			if left == "rates" {
				currencies := parseData["rates"].(map[string]interface{})
				rateRub := currencies["RUB"].(float64)
				rateCur := currencies[currency].(float64)
				return balance / rateRub * rateCur
			}
		}
	}
	return -1.0
}

func readData(sign float64, dec *json.Decoder, db *sql.DB, w http.ResponseWriter) {

	var req TransferRequest
	if !checkErr(dec.Decode(&req), w) {
		id, err := strconv.Atoi(req.UserID)
		if checkErr(err, w) || id <= 0  {
			http.Error(w, "Id should be positive integer!", http.StatusNotFound)
			return
		}
		sum, err := strconv.ParseFloat(req.Sum, 64)
		if checkErr(err, w) || sum <= 0 {
			http.Error(w, "Sum should have positive value!", http.StatusNotFound)
			return
		}
		changeBankAccount(sum, sign, id, db, w, -1, req.Info)
	}
}

func getReport(field string, table string, dec *json.Decoder, db *sql.DB, w http.ResponseWriter, sortBy string ) {

	var user UserID
	var sum, finalBalance float64
	var created_at, info, req string
	var fromOrToID int
	if checkErr(dec.Decode(&user), w) {
		return
	}
	id, err := strconv.Atoi(user.UserID)
	if checkErr(err, w) || id <= 0  {
		http.Error(w, "Id should be positive integer!", http.StatusNotFound)
		return
	}
	if sortBy!="" {
		req = "select " + field + ", sum, info, finalBalance, created_at from " + table + " where user = ? order by " + sortBy + " desc;"
	} else {
		req = "select " + field + ", sum, info, finalBalance, created_at from " + table + " where user = ? ;"
	}

	stmt, err := db.Query(req, id)
	if !checkErr(err, w) {
		sendResponse(w, fmt.Sprintf("Report for user with id = %d with columns (fromOrToID, sum, info, finalBalance, created_at)", id))

		for stmt.Next() {
			err := stmt.Scan(&fromOrToID, &sum, &info, &finalBalance, &created_at)
			checkErr(err, w)
			sendResponse(w, fmt.Sprintf("%d %.2f %s %.2f %s", fromOrToID, sum, info, finalBalance, created_at))
		}
		stmt.Close()
	}
}

func changeBankAccount(sum float64, sign float64, id int, db *sql.DB, w http.ResponseWriter, fromOrToID int, info string) bool {

	var balance float64
	stmt1, err1 := db.Prepare("update user set balance = ? where id = ? ;")
	stmt2, err2 := db.Query("select balance from user where id = ? ;", id)
	if !checkErr(err1, w) && !checkErr(err2, w) {                 //changing user balance
		if stmt2.Next() {
			err := stmt2.Scan(&balance)
			checkErr(err, w)
		} else {
			http.Error(w, fmt.Sprintf("There is no user with id %d!", id), http.StatusNotFound)
			return true
		}
		stmt2.Close()
		if balance+sign*sum < 0 {
			http.Error(w, "Balance can't be negative!", http.StatusNotFound)
			return true                                                                   //return error
		}
		_, err := stmt1.Exec(balance+sign*sum, id)
		if checkErr(err, w) {
			return true
		}
		if sign > 0 {
			stmt1, err1 = db.Prepare("insert into charge(sum, user, fromID, finalBalance, info) values( ? , ? , ? , ? , ? );")
		} else {
			stmt1, err1 = db.Prepare("insert into writeOff(sum, user, toID, finalBalance, info) values( ? , ? , ? , ? , ? );")
		}
		checkErr(err1, w)
		_, err1 = stmt1.Exec(sum, id, fromOrToID, balance+sign*sum, info)
		if !checkErr(err1, w) {
			return false
		}
		return true
	} else {
		return true
	}
}

func requestsHandler(w http.ResponseWriter, req *http.Request) {
	db, err := sql.Open("mysql", "user1:password1@tcp(172.22.0.2:3306)/testdb")   //docker inspect golang_db_avito | grep IPAddr
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatal(err.Error())
	}
	defer db.Close()

	if req.Method == "POST"{
		dec := checkHeader(w, req)
		if dec != nil {
			if req.URL.Path == "/users/add" {

				var user User
				checkErr(dec.Decode(&user), w)

				stmt, err := db.Prepare("insert into user(username) values(?)")
				checkErr(err, w)

				res, err := stmt.Exec(user.Username)
				if !checkErr(err, w) {
					id, err := res.LastInsertId()
					checkErr(err, w)
					sendResponse(w, fmt.Sprintf("%s %d", "Created new user with id =", id))
				}

			} else if req.URL.Path == "/users/charge" {

				readData(1.0, dec, db, w)

			} else if req.URL.Path == "/users/writeOff" {

				readData(-1.0, dec, db, w)

			} else if req.URL.Path == "/users/transfer" {

				var req TransferFromToRequest
				checkErr(dec.Decode(&req), w)

				id1, err1 := strconv.Atoi(req.FromID)
				id2, err2 := strconv.Atoi(req.ToID)
				if checkErr(err1, w) || checkErr(err2, w) || id1 <= 0 || id2 <= 0  {
					http.Error(w, "Id should be positive integer!", http.StatusNotFound)
					return
				}
				if id1 == id2 {
					http.Error(w, "You can't tranfer money to your account!", http.StatusNotFound)
					return
				}
				sum, _ := strconv.ParseFloat(req.Sum, 64)
				if checkErr(err, w) || sum <= 0 {
					http.Error(w, "Sum should have positive value!", http.StatusNotFound)
					return
				}

				//если нет отправителя, то ничего не произойдёт. Если отправитель есть, но нет получателя, то деньги со счёта отправителя сначала снимутся,
				//а потом вернутся с соответствующим комментарием

				if !changeBankAccount(sum, -1.0, id1, db, w, id2, req.Info) {
					if changeBankAccount(sum, 1.0, id2, db, w, id1, req.Info) {
						changeBankAccount(sum, 1.0, id1, db, w, -1, req.Info+". No getter. Money returned back")
					}
				}

			} else if req.URL.Path == "/users/getBalance" {

				var user UserID
				var balance float64
				checkErr(dec.Decode(&user), w)
				id, err := strconv.Atoi(user.UserID)
				if checkErr(err, w) || id <= 0  {
					http.Error(w, "Id should be positive integer!", http.StatusNotFound)
					return
				}

				stmt, err := db.Query("select balance from user where id = ? ;", id)
				if !checkErr(err, w) {
					if stmt.Next() {
						err := stmt.Scan(&balance)
						checkErr(err, w)
					} else {
						http.Error(w, fmt.Sprintf("There is no user with id %d!", id), http.StatusNotFound)
						return
					}
					stmt.Close()
				} else {
					return
				}

				sendResponse(w, fmt.Sprintf("Balance of user with id = %d is %.2f", id, balance))

				if cur:=req.URL.Query().Get("currency"); cur!="" {                                     //returns empty string if not found
					convertedValue := getRate(cur, balance, w)
					sendResponse(w, fmt.Sprintf("%.2f RUB in %s is %.2f", balance, cur, convertedValue))
				}

			} else if req.URL.Path == "/users/getChargeReport" {

				cur:=req.URL.Query().Get("sort");

				if cur=="date" {
					getReport("fromID", "charge", dec, db, w, "created_at")
				} else if cur=="sum" {
					getReport("fromID", "charge", dec, db, w, "sum")
				} else {
					getReport("fromID", "charge", dec, db, w, "")
				}

			} else if req.URL.Path == "/users/getWriteOffReport" {

				cur:=req.URL.Query().Get("sort");

				if cur=="date" {
					getReport("toID", "writeOff", dec, db, w, "created_at")
				} else if cur=="sum" {
					getReport("toID", "writeOff", dec, db, w, "sum")
				} else {
					getReport("toID", "writeOff", dec, db, w, "")
				}
			} else {
				http.Error(w, "URL path is not found.\n", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "Body is empty.\n", http.StatusNoContent)
		}
	} else {
		http.Error(w, "Method is not supported.\n", http.StatusNotFound)
		return
	}
}

func main() {
	http.HandleFunc("/", requestsHandler)
	fmt.Printf("Starting server at port 9000\n")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
