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
/tmp/facebook_db.bin
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

### query

```
$ sqlite3 /tmp/facebook_db.bin
sqlite>
```

댓글 많이 달린 글 상위 10

```
select A.*, feeds.* from feeds inner join
(select count(*) as cnt, feed_id from comments group by feed_id) as A
on feeds.id = A.feed_id order by cnt desc limit 10;
```

댓글 많이 달린 글 상위 10 (날짜 조건 추가)

```
select A.*, feeds.* from feeds inner join
(select count(*) as cnt, feed_id from comments group by feed_id) as A
on feeds.id = A.feed_id where feeds.created_at >= strftime('%s', '2015-01-01')
order by cnt desc limit 10;
```


포스트 많이 한 사람 상위 10

```
select users.name, cnt from users inner join
(select count(*) cnt, "from" from feeds group by "from")
as A
on users.id = A."from" order by cnt desc limit 10;
```

포스트 많이 한 사람 상위 10 (날짜 조건 추가)

```
select users.name, cnt from users inner join
(select count(*) cnt, "from" from feeds where updated_at >= strftime('%s',
'2015-01-01') group by "from") as A
on users.id = A."from" order by cnt desc limit 10;
```

댓글 많이 한 사람 상위 10

```
select users.name, cnt from users inner join
(select count(*) cnt, "from" from comments group by "from") as A
on id = A."from" order by cnt desc limit 10;
```

댓글 많이 한 사람 상위 10 (날짜 조건 추가)

```
select users.name, cnt from users inner join
(select count(*) cnt, "from" from comments where created_at >= strftime('%s', '2015-01-01') 
group by "from") as A
on id = A."from" order by cnt desc limit 10;
```

like 많이 한 사람 상위 10

```
select users.name, cnt from users inner join
(select count(*) cnt, user_id from likes group by user_id) as A
on users.id = A.user_id order by cnt desc limit 10;
```
