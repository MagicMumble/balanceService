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
}

type TransferFromToRequest struct {
	FromID string `json:"fromID"`
	ToID string `json:"toID"`
	Sum    string `json:"sum"`
}

type UserID struct {
	UserID string `json:"userID"`
}

type Response struct {
	Response string `json:"Response"`
}

func sendResponse(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "application/json")        //вернёт id нового пользователя
	data := Response {
		Response: text,
	}
	json.NewEncoder(w).Encode(data)
}

func checkHeader(w http.ResponseWriter, req *http.Request) *json.Decoder {
	contType := req.Header.Get("Content-Type")
	if contType != "" && contType != "application/json" {
		msg := "Content-Type header is not application/json"
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return nil
	}
	req.Body = http.MaxBytesReader(w, req.Body, 1048576)
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	return dec
}

func checkErr(err error, w http.ResponseWriter, message string) bool {
	if err != nil {
		log.Println(err)                                  //to server
		sendResponse(w, message)
		return true
	} else {
		return false
	}
}

func getRate(currency string, balance float64) float64{
	response, err := http.Get("http://openexchangerates.org/api/latest.json?app_id="+APP_ID+"&symbols=RUB,"+currency) //would just change base but this option isn't for free :(
	checkErr(err)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	checkErr(err)

	parseData := make(map[string]interface{})
	err = json.Unmarshal([]byte(string(body)), &parseData)
	checkErr(err)

	for left, _ := range parseData {
		if left == "rates" {
			currencies := parseData["rates"].(map[string]interface{})
			rateRub := currencies["RUB"].(float64)
			rateCur := currencies[currency].(float64)
			return balance/rateRub*rateCur
		}
	}
	return -1.0
}

func readData(sign float64, dec *json.Decoder, db *sql.DB, w http.ResponseWriter) {
	var req TransferRequest
	checkErr(dec.Decode(&req))

	id, _ := strconv.Atoi(req.UserID)                   //ошибки проверяются на уровне бд
	sum, _ := strconv.ParseFloat(req.Sum, 64)

	changeBankAccount(sum, sign, id, db, w, -1)
}

func getReport(field string, table string, dec *json.Decoder, db *sql.DB, w http.ResponseWriter, sortBy string ) {

	var user UserID
	var sum, finalBalance float64
	var created_at, info, req string
	var fromOrToID int
	checkErr(dec.Decode(&user))
	id, _ := strconv.Atoi(user.UserID)
	if sortBy!="" {
		req = "select "+field+", sum, info, finalBalance, created_at from "+table+" where user = ? order by "+sortBy+ " desc;"
	} else {
		req = "select "+field+", sum, info, finalBalance, created_at from "+table+" where user = ? ;"
	}

	stmt, err := db.Query(req, id)
	if !checkErr(err) {
		fmt.Println("Report for user with id = ", id, ":")
		for stmt.Next() {
			err := stmt.Scan(&fromOrToID, &sum, &info, &finalBalance, &created_at)
			checkErr(err)
			fmt.Println(fromOrToID, sum, info, finalBalance, created_at)
			sendResponse(w, fmt.Sprintf("%d %.2f %s %.2f %s", fromOrToID, sum, info, finalBalance, created_at))
		}
		stmt.Close()
	}
}

func changeBankAccount(sum float64, sign float64, id int, db *sql.DB, w http.ResponseWriter, fromOrToID int) bool {

	var balance float64

	stmt1, err1 := db.Prepare("update user set balance = ? where id = ? ;")
	stmt2, err2 := db.Query("select balance from user where id = ? ;", id)
	if !checkErr(err1) && !checkErr(err2) { //изменили баланс пользователя
		if stmt2.Next() {
			err := stmt2.Scan(&balance)
			checkErr(err)
		}
		stmt2.Close()
		if balance+sign*sum < 0 {
			log.Println("Balance can't be negative")
			return true //error!
		}
		_, err := stmt1.Exec(balance+sign*sum, id)
		if !checkErr(err) {
			fmt.Println("Charged bank account for user with id = ", id)
			sendResponse(w, strconv.Itoa(id))
		}
		if sign > 0 {
			stmt1, err1 = db.Prepare("insert into charge(sum, user, fromID, finalBalance) values( ? , ? , ? , ? );")
		} else {
			stmt1, err1 = db.Prepare("insert into writeOff(sum, user, toID, finalBalance) values( ? , ? , ? , ? );")
		}
		checkErr(err1)
		_, err1 = stmt1.Exec(sum, id, fromOrToID, balance+sign*sum)
		checkErr(err1)
		return false               //то есть ошибок нет
	} else {
		return true
	}
}

func requestsHandler(w http.ResponseWriter, req *http.Request) {
	db, err := sql.Open("mysql", "user1:password1@/testdb")
	checkErr(err)
	defer db.Close()

	if req.Method == "POST"{
		dec := checkHeader(w, req)
		if dec != nil {
			if req.URL.Path == "/users/add" {

				var user User
				checkErr(dec.Decode(&user))

				stmt, err := db.Prepare("insert into user(username) values(?)")
				checkErr(err)

				res, err := stmt.Exec(user.Username)
				if !checkErr(err) {
					id, err := res.LastInsertId()
					checkErr(err)
					fmt.Println("Created new user with id = ", id)
					sendResponse(w, strconv.Itoa(int(id)))
				}

			} else if req.URL.Path == "/users/charge" {

				readData(1.0, dec, db, w)

			} else if req.URL.Path == "/users/writeOff" {

				readData(-1.0, dec, db, w)

			} else if req.URL.Path == "/users/transfer" {

				var req TransferFromToRequest
				checkErr(dec.Decode(&req))

				id1, err1 := strconv.Atoi(req.FromID)
				id2, err2 := strconv.Atoi(req.ToID)
				if !checkErr(err1) && !checkErr(err2) {
					if id1 == id2 {
						log.Println("You can't tranfer money to your account!")
						return
					}
					sum, _ := strconv.ParseFloat(req.Sum, 64)

					if !changeBankAccount(sum, -1.0, id1, db, w, id2) {
						changeBankAccount(sum, 1.0, id2, db, w, id1)
					}
				}

			} else if req.URL.Path == "/users/getBalance" {

				var user UserID
				var balance float64
				checkErr(dec.Decode(&user))
				id, _ := strconv.Atoi(user.UserID)

				stmt, err := db.Query("select balance from user where id = ? ;", id)
				if !checkErr(err) {
					if stmt.Next() {
						err := stmt.Scan(&balance)
						checkErr(err)
					}
					stmt.Close()
				}

				fmt.Println("Balance of user with id = ", id, " is ", balance)
				if cur:=req.URL.Query().Get("currency"); cur!="" {                       //returns empty string if not found
					convertedValue := getRate(cur, balance)
					fmt.Println("Converted in "+cur+" = ", convertedValue)
					sendResponse(w, fmt.Sprintf("%.2f in %s is %.2f", balance, cur, convertedValue))
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
				http.Error(w, "404 not found.\n", http.StatusNotFound)
				return
			}
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
