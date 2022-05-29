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

### Authentication

I have implemented a basic authentication middleware. All requests must have the `X-Token` header with `test` value.



## TO DO

* Unit tests

## Known issues

* The query `SELECT * FROM logs LIMIT n` in `count()` method makes `rows.Next()` to panic. Because of this, the request `GET /?n=10` (or any number) will make the application to crash.
* Test coverage could be better :-)

> coverage: 61.2% of statements
> 
> ok  	github.com/robertolopezlopez/immudemo	1.225s	coverage: 61.2% of statements
