# Golang / Mgo file server

### Example
curl:
```sh
curl -X POST \
  'http://localhost/url' \
  -H 'Authorization: Basic dGVzdDp0ZXN0' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data 'uploadfile=http%3A%2F%2Fsct.ru%2F_nuxt%2Fimg%2Flogo.b13867c.png&pngqlt=60&jpgqlt=75' \
  --compressed
```

response:
```json
{
  "ID": "5a1d87f73a7e65000b685f7f",
  "Name": "logo.b13867c.png",
  "Path": "./store/f78bb4352b7f0f1cac79fe204f07f69e-logo.b13867c.png",
  "ContentType": "image/png",
  "InSize": 7237,
  "OutSize": 1379,
  "UserName": "test",
  "Host": "sct.ru",
  "Keep": 1,
  "Timestamp": "2017-11-28T18:59:51.965003054+03:00"
}
```

Last 500 file links @ http://localhost/

File stored @ http://localhost/store/59e5ab624cd0a60009b0f59f

User statistic @ http://localhost/stat

### Enviroment variables (docker-compose.yml app.enviroment)
```yml
HOST=0.0.0.0
PORT=8080
TEST_QUANTITY=100
DAYS_TO_KEEP=90
TZ=Europe/Moscow
```

users:
```json
[
  {
    "name": "test",
    "password": "test"
  }
]
```

To update users run:
```sh
docker-compose build seed && docker-compose run seed
```
