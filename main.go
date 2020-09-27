package main
import (
	"database/sql"
	"math"

	//"encoding/base64"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	//"strings"
)
var db *sql.DB
var err error
var eur float32
var date string
type Request struct{
	Id1 int 	`json:"id1"`
	Id2 int 	`json:"id2"`
	Count int 	`json:"cnt"`
}
type Response struct{
	Message string `json:"msg"`
}

func getConvertMoney(balance int, to string) string {
	var to_http string  = ","
	if (to != "EUR"){
		to_http += to
	} else {
		to_http = ""
	}

	resp, _ := http.Get("https://api.exchangeratesapi.io/latest?symbols=RUB"+to_http)
	defer resp.Body.Close()
	var result map[string]map[string] float64

	err = json.NewDecoder(resp.Body).Decode(&result)

	var eur_rub float64 = result["rates"]["RUB"]
	var eur_smth float64 = result["rates"][to]
	var res float64 = float64(balance)/(eur_rub/eur_smth)
	if (to_http == ""){
		to = "EUR"
		res = float64(balance)/eur_rub
	}
	if (math.IsNaN(res)){
		return strconv.Itoa(balance) + "RUB"
	}
	return strconv.FormatFloat(res,'f',2,64) + to
}

func addMoney(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	var req Request
	var res Response
	_ = json.NewDecoder(r.Body).Decode(&req)
	if (req.Id1 <= 0 || req.Count < 0){
		res.Message = "Error. id must be more than 0 and cnt >= 0"
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	_, err = db.Exec("INSERT INTO balance_schema.balances VALUES (?, ?) ON DUPLICATE KEY UPDATE Balance = Balance + ?" , req.Id1, req.Count, req.Count)
	if (err != nil){
		res.Message = "Error"
		panic(err)
	} else {
		res.Message = "OK."
	}
	_ = json.NewEncoder(w).Encode(res)
}
func withdrawMoney(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	var req Request
	var res Response
	_ = json.NewDecoder(r.Body).Decode(&req)
	if (req.Id1 <= 0) || (req.Count <=0){
		res.Message = "Error. id and cnt must be more than 0"
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	rows, err := db.Query("SELECT balance FROM balance_schema.balances WHERE Id = ?" , req.Id1)
	var balance int = -1
	for rows.Next() {
		rows.Scan(&balance)
	}
	if (balance <= 0 || balance < req.Count){
		res.Message = "Error. not balance enough"
		_ = json.NewEncoder(w).Encode(res)
		return
	} else {
		balance -= req.Count
		_, err = db.Exec("UPDATE balance_schema.balances SET Balance = ? WHERE Id = ?", balance, req.Id1)
	}
	if (err != nil){
		res.Message = "Error"
		panic(err)
	} else {
		res.Message = "OK."
	}
	_ = json.NewEncoder(w).Encode(res)
}

func tradeMoney(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req Request
	var res Response
	_ = json.NewDecoder(r.Body).Decode(&req)
	if (req.Id1 <= 0) || (req.Count <= 0) || (req.Id2 <= 0) {
		res.Message = "Error. id1 and id2 and cnt must be more than 0"
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	rows, _ := db.Query("SELECT balance FROM balance_schema.balances WHERE Id = ?" , req.Id1)
	var balance int = -1
	for rows.Next() {
		rows.Scan(&balance)
	}
	if (balance <= 0 || balance < req.Count){
		res.Message = "Error.2 not balance enough"
		_ = json.NewEncoder(w).Encode(res)
		return
	} else {
		balance -= req.Count
		_, err = db.Exec("UPDATE balance_schema.balances SET Balance = ? WHERE Id = ?", balance, req.Id1)
		if (err == nil){
			_, err = db.Exec("INSERT INTO balance_schema.balances VALUES (?, ?) ON DUPLICATE KEY UPDATE Balance = Balance + ?" , req.Id2, req.Count, req.Count)
		}
	}
	if (err != nil){
		res.Message = "Error"
		panic(err)
	} else {
		res.Message = "OK."
	}
	_ = json.NewEncoder(w).Encode(res)
}
func getMoney(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req Request
	var res Response
	_ = json.NewDecoder(r.Body).Decode(&req)
	if (req.Id1 <= 0 ){
		res.Message = "Error. id must be more than 0"
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	rows, err := db.Query("SELECT balance FROM balance_schema.balances WHERE Id = ?" , req.Id1)
	var balance int = -1
	for rows.Next() {
		rows.Scan(&balance)
	}
	if (err != nil || balance < 0){
		res.Message = "Error"
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	params := mux.Vars(r)
	var sbalance string= strconv.Itoa(balance) + "RUB"
	if (params["currency"] != ""){
		sbalance = getConvertMoney(balance, params["currency"])
	}
	res.Message = "User id: " + strconv.Itoa(req.Id1) + " balance: " + sbalance
	_ = json.NewEncoder(w).Encode(res)
}

func main() {
	db, err = sql.Open("mysql", "root:Root2000@/balance_schema")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err != nil{
		panic(err)
	}

	fmt.Println("Server is running...")
	r := mux.NewRouter()
	r.HandleFunc("/add", addMoney).Methods("POST")
	r.HandleFunc("/withdraw", withdrawMoney).Methods("POST")
	r.HandleFunc("/trade", tradeMoney).Methods("POST")
	r.HandleFunc("/{currency}", getMoney).Methods("POST")
	r.HandleFunc("/", getMoney).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", r))
}
