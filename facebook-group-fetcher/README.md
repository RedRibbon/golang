# Facebook redribbon group feeds fetcher


## How to run

```
$ go get github.com/RedRibbon/golang/facebook-group-fetcher
$ facebook-group-fetcher group -t token.dat
```

`token.dat` is AccessToken file to access redribbon facebook group. 
It is not included in this REPO.

## database

database는 `sqlite3`를 사용하고 파일경로는 다음과 같다.

```
/tmp/facebook_db.bin`
```

fetcher를 실행하면 기존 데이타를 모두 지우고 시작한다.


### ERD

```
-----------------
| feeds         |
-----------------
| id            |
| from          |
| message       |
| created_at    |
| updated_at    |
-----------------
```

```
-----------------
| comments      |
-----------------
| id            |
| feed_id       |
| from          |
| message       |
| line_count    |
| created_at    |
-----------------
```

```
-----------------
| likes         |
-----------------
| id            |
| user_id       |
| feed_id       |
-----------------
```

```
-----------------
| user          |
-----------------
| id            |
| name          |
-----------------
```
