# immudemo

## docker compose

* immudb
  * port: 3322
  * web interface port: 8081
  * User: `immudb`
  * Password: `immudb`
* app
  * port: 8080

## Endpoints

* `GET /`: show logs in reverse chronological order
  * Optional query parameters:
    * n: int, number of rows
* `GET /count`: return number of stored logs
* `POST /`: log messages in a transaction
  * JSON body format:
    * `msgs`: array of messages to log, must not be empty

## TO DO

* Unit tests
* Basic authentication middleware

## known issues

* The query `SELECT * FROM logs LIMIT n` in `count()` method makes `rows.Next()` to panic.