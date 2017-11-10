# Golang / Mgo file server

### Example
curl:
```sh
curl -X POST \
  'http://77.244.214.132/url' \
  -H 'Authorization: Basic dGVzdDp0ZXN0' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data 'uploadfile=http%3A%2F%2Fsct.ru%2F_nuxt%2Fimg%2Flogo.b13867c.png&pngqlt=60&jpgqlt=75' \
  --compressed
```

response:
```json
{
  "ID": "59e5ab624cd0a60009b0f59f",
  "Name": "stock-photo-fairy-path-140700811-1600px-1500x1000.jpg",
  "Path": "./store/226b0d28460e12bc381ee63405c3f8e6-stock-photo-fairy-path-140700811-1600px-1500x1000.jpg",
  "ContentType": "image/jpeg",
  "InSize": 299710,
  "OutSize": 184094,
  "User": "test2",
  "Host": "iso.500px.com",
  "Keep": 0,
  "Timestamp": "2017-10-17T10:04:02.388412239+03:00"
}
```

File stored at http://localhost/store/59e5ab624cd0a60009b0f59f

Last 500 file links at http://localhost/store

### Enviroment variables (docker-compose.yml app.enviroment)
```yml
HOST=0.0.0.0
PORT=8080
TEST_QUANTITY=1000
DAYS_TO_KEEP=1
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
docker-compose build seed && docker-compose up seed
```
