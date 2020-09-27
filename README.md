# BalanceService
Микросервис для работы с балансом пользователей. На вход принимается HTTP запрос с JSON файлом для получения баланса, для начисления, списания
или обмена деньгами пользователей. В качестве СУБД была выбрана MySQL.
## Настройка БД
Чтобы создать нужную таблицу базы данных необходимо выполнить следующий код в SQL Editor:
```sql
CREATE TABLE `balances` (
  `Id` int NOT NULL,
  `Balance` int DEFAULT NULL,
  PRIMARY KEY (`Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
SELECT * FROM balance_schema.balances;
```
Также необходимо в коде main заменить строку (пароль и имя своей таблицы)
```go
db, err = sql.Open("mysql", "root:<password>@/<table_name>")
```
### Получение баланса пользователя
1.  Пример запроса: `http://localhost:8000/, POST, JSON: { "id1" :1 }`
    Ответ: JSON `{ "msg": "User id: 1 balance: 173RUB" }`
2.  Пример запроса:`http://localhost:8000/USD, POST, JSON: { "id1" :1 }`
    Ответ: JSON `{ "msg": "User id: 1 balance: 2.23USD" }`
В первом случае происходил запрос в рублях, во втором в долларах. Конвертация с помощью сайта из ТЗ https://exchangeratesapi.io/
### Зачисление средств пользователю
    Пример запроса: http://localhost:8000/add, POST, JSON: { "id1" :1 , "cnt" : 99}
    Ответ: JSON { "msg": "OK" }
Создание нового пользователя происходит при первом зачислении средств
### Списание средств пользователя
1.    Пример запроса: `http://localhost:8000/withdraw, POST, JSON: { "id1" :1 , "cnt" : 98}`
      Ответ: `JSON { "msg": "OK" }`
2.    Пример запроса: `http://localhost:8000/withdraw, POST, JSON: { "id1" :1 , "cnt" : 98}`
      Ответ: `JSON { "msg": "Error. not balance enough" }`
При недостаточном балансе или неправильном id пользователя присылается ошибка
 ### Обмен средствами между пользователями
      Пример запроса: `http://localhost:8000/, POST, JSON: { "id1" :1 , "id2" :2, "cnt" :123 }`
      Ответ: JSON `{ "msg": "OK" }`
При недостаточном балансе первого пользователя или неверном ID выведется сообщение об ошибке
