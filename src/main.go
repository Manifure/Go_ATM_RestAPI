package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Интерфейс банковского счета
type BankAccount interface {
	Deposit(amount float64) error
	Withdraw(amount float64) error
	GetBalance() float64
}

// Структура аккаунта
type Account struct {
	ID      int     `json:"id"`
	Balance float64 `json:"balance"`
	mutex   sync.Mutex
}

// Реализация методов интерфейса BankAccount для структуры Account
func (a *Account) Deposit(amount float64) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.Balance += amount
	log.Printf("[%s] Deposit: Account %d, Amount: %.2f, New Balance: %.2f\n", time.Now().Format(time.RFC3339), a.ID, amount, a.Balance)
	return nil
}

func (a *Account) Withdraw(amount float64) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.Balance >= amount {
		a.Balance -= amount
		log.Printf("[%s] Withdrawal: Account %d, Amount: %.2f, New Balance: %.2f\n", time.Now().Format(time.RFC3339), a.ID, amount, a.Balance)
		return nil
	} else {
		log.Printf("[%s] Insufficient funds: Account %d, Balance: %.2f, Requested Amount: %.2f\n", time.Now().Format(time.RFC3339), a.ID, a.Balance, amount)
		return fmt.Errorf("insufficient funds")
	}
}

func (a *Account) GetBalance() float64 {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	log.Printf("[%s] Balance check: Account %d, Balance: %.2f\n", time.Now().Format(time.RFC3339), a.ID, a.Balance)
	return a.Balance
}

// Для простоты я использовал мапу, в идеале нужна база данных
var accounts = make(map[int]*Account)
var nextAccountID = 1
var mutex sync.Mutex

// Обработчик создания аккаунта
func createAccountHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	account := &Account{
		ID:      nextAccountID,
		Balance: 0,
	}
	accounts[nextAccountID] = account
	nextAccountID++

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
	log.Printf("[%s] Account created: ID %d\n", time.Now().Format(time.RFC3339), account.ID)
}

// Обработчик пополнения баланса
func depositHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	accountID := 0
	fmt.Sscanf(id, "%d", &accountID)

	if account, ok := accounts[accountID]; ok {
		var requestBody struct {
			Amount float64 `json:"amount"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		go func() {
			account.Deposit(requestBody.Amount)
		}()

		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Account not found", http.StatusNotFound)
	}
}

// Обработчик снятия средств
func withdrawHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	accountID := 0
	fmt.Sscanf(id, "%d", &accountID)

	if account, ok := accounts[accountID]; ok {
		var requestBody struct {
			Amount float64 `json:"amount"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		go func() {
			err := account.Withdraw(requestBody.Amount)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}()

		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Account not found", http.StatusNotFound)
	}
}

// Обработчик проверки баланса
func balanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	accountID := 0
	fmt.Sscanf(id, "%d", &accountID)

	if account, ok := accounts[accountID]; ok {
		balance := account.GetBalance()
		response := map[string]interface{}{"balance": balance}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Account not found", http.StatusNotFound)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/accounts", createAccountHandler).Methods("POST")
	router.HandleFunc("/accounts/{id}/deposit", depositHandler).Methods("POST")
	router.HandleFunc("/accounts/{id}/withdraw", withdrawHandler).Methods("POST")
	router.HandleFunc("/accounts/{id}/balance", balanceHandler).Methods("GET")

	fmt.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
