# Go_ATM_RestAPI
Тестовое задание, в котором необходимо реализовать REST API на Golang, которое имитирует работу банкомата.
на выполнение задания даётся 24 часа.

# Пример использования

## Создание нового аккаунта
curl -X POST http://localhost:8080/accounts
### {"id":1,"balance":0}

## Пополнение баланса аккаунта 1 на 100
curl -X POST -H "Content-Type: application/json" -d '{"amount": 100}' http://localhost:8080/accounts/1/deposit

## Проверка баланса аккаунта 1
curl http://localhost:8080/accounts/1/balance
### {"balance":100}

## Снятие средств с аккаунта 1 на 12
curl -X POST -H "Content-Type: application/json" -d '{"amount": 12}' http://localhost:8080/accounts/1/withdraw
