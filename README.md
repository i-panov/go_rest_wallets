# 💰 Go REST Wallets

Простой REST API для управления виртуальными кошельками с поддержкой PostgreSQL, Docker и тестами.

## 📌 Основные возможности

- Получение информации о кошельке по ID
- Пополнение баланса
- Снятие средств
- Автоматическое создание новых кошельков:
  - при вызове метода `POST /api/v1/wallet` с `operationType: "DEPOSIT"` и пустым `walletId`
  - при попытке пополнить несуществующий кошелёк
- Поддержка высокой нагрузки — более **7000 RPS**
- Unit-тесты и нагрузочное тестирование через `wrk`

---

## 🧩 Технологии

| Технология | Использование |
|------------|----------------|
| Go         | 1.23           |
| PostgreSQL | Хранение данных |
| pgx        | Драйвер БД     |
| gorilla/mux| Роутинг        |
| testify    | Для unit-тестов |
| Docker     | Контейнеризация |

---

## 🔧 Установка и запуск

### 1. Создайте `.env` файл

```bash
cp .env.example .env
```

Укажите правильные значения.

---

## 🚀 Запуск через Docker (рекомендуется)

```bash
docker-compose up -d
```

API будет доступен на порту `8080`.

---

## 🧪 Unit-тесты

Тесты находятся в папке `src/tests`.  
Проверяют логику репозитория, обработку ошибок и валидацию входных данных.

### Как запустить:

```bash
go test -v ./src/tests
```

---

## 🧱 Load Testing с помощью `wrk`

Для нагрузочного тестирования используется утилита [`wrk`](https://github.com/wg/wrk ).  
Файл `wrk.lua` задаёт параметры POST-запроса (`Content-Type`, `body`).

### ✅ GET `/api/v1/wallets/{id}` — получение кошелька

```bash
$ wrk -t12 -c1000 -d30s http://localhost:8080/api/v1/wallets/14f21088-2613-430d-b412-f82ce3559dcd
Running 30s test @ http://localhost:8080/api/v1/wallets/14f21088-2613-430d-b412-f82ce3559dcd
  12 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    31.34ms    5.25ms  92.68ms   88.18%
    Req/Sec     2.66k   375.77    10.25k    83.23%
  954443 requests in 30.08s, 159.29MB read
Requests/sec:  31729.54
Transfer/sec:      5.30MB
```

---

### ✅ POST `/api/v1/wallet` — пополнение/снятие

```bash
$ wrk -t12 -c1000 -d30s -s wrk.lua http://localhost:8080/api/v1/wallet
Running 30s test @ http://localhost:8080/api/v1/wallet
  12 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   135.60ms   17.65ms 312.96ms   91.08%
    Req/Sec   614.85    142.06     1.24k    68.19%
  220220 requests in 30.09s, 36.75MB read
Requests/sec:   7318.56
Transfer/sec:      1.22MB
```

> Сервис успешно держит нагрузку:  
> ✅ **GET** — до **31,729 RPS**  
> ✅ **POST** — до **7,318 RPS**

---

## 🧪 Примеры использования

### ✅ Получить кошелёк

```bash
curl http://localhost:8080/api/v1/wallets/a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11
```

### ✅ Пополнить существующий кошелёк

```bash
curl -X POST http://localhost:8080/api/v1/wallet \
     -H "Content-Type: application/json" \
     -d '{"walletId":"existing-id","operationType":"DEPOSIT","amount":50}'
```

### ✅ Снять средства

```bash
curl -X POST http://localhost:8080/api/v1/wallet \
     -H "Content-Type: application/json" \
     -d '{"walletId":"existing-id","operationType":"WITHDRAW","amount":30}'
```

### ✅ Создать новый кошелёк

```bash
curl -X POST http://localhost:8080/api/v1/wallet \
     -H "Content-Type: application/json" \
     -d '{"operationType":"DEPOSIT","amount":100}'
```

> При пустом или несуществующем `walletId` — кошелёк создаётся автоматически.

---

## 🧪 Примеры ошибок

### ❌ Кошелёк не найден

```json
{
  "error": "Wallet not found"
}
```

### ❌ Валидационная ошибка

```json
{
  "error": "Validation error on field 'id': cannot be empty"
}
```
