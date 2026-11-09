# 03. MySQL과 데이터 접근

## database/sql과 드라이버

`database/sql`은 Go 표준 라이브러리의 DB 접근 API다. MySQL 통신은 `go-sql-driver/mysql`이 맡는다. `sql.Open`은 핸들을 만들며 실제 연결 성공을 보증하지 않는다. 시작할 때 `PingContext`로 확인하는 이유다.

`*sql.DB`는 연결 하나가 아니라 **연결 풀**이다. 서버 시작 시 한 번 만들고 요청들에서 공유한다. 요청마다 Open과 Close를 하지 않는다. `MaxOpenConns`는 동시 열린 연결 상한이고, `MaxIdleConns`는 재사용을 위해 남길 유휴 연결 상한이다. 풀이 작으면 연결 대기가, 너무 크면 DB 부하가 늘 수 있다. `db.Stats()`의 InUse, Idle, WaitCount, WaitDuration을 함께 본다.

예제의 최대 10개는 로컬 학습 설정이다. 실제 값은 인스턴스 수를 곱한 전체 연결 예산과 DB 처리 능력을 보고 정한다. `parseTime=true`는 DATETIME/TIMESTAMP를 time.Time으로 읽는 데 필요하고 `loc=UTC`는 드라이버의 시간 해석 기준이다. DB 세션의 timezone 설정과 같은 옵션은 아니다.

## Query, QueryRow, Exec

| API | 사용 | 반드시 볼 것 |
| --- | --- | --- |
| QueryContext | 여러 행 조회 | Rows.Close, Next, Scan, Rows.Err |
| QueryRowContext | 한 행 조회 | Scan 시점 오류, sql.ErrNoRows |
| ExecContext | INSERT/UPDATE/DELETE | 오류와 RowsAffected |

`?`로 값을 바인딩한다. 사용자 입력을 문자열 연결로 SQL에 넣지 않는다. 테이블명이나 ORDER BY 컬럼은 값 placeholder로 대체할 수 없으므로 허용 목록에서 고른다.

```go
row := db.QueryRowContext(ctx, "SELECT title FROM reading_lists WHERE id = ?", id)
err := row.Scan(&title)
```

## 스키마를 고른 이유

목록과 항목을 나눠 목록 하나에 자료 여러 개를 둔다. 항목의 PK `(list_id, resource_id)`는 같은 목록의 중복 자료를 막는다. `(list_id, position)` UNIQUE는 순서 중복을 막고 목록별 정렬 조회에도 사용할 수 있다. request_key는 UNIQUE로 동일 생성 요청을 식별한다.

InnoDB의 PK는 clustered index를 구성하고 secondary index에도 PK 값이 들어간다. 무작정 큰 PK나 인덱스를 많이 두면 저장 공간과 쓰기 비용이 커진다. `EXPLAIN`으로 실제 조회 계획을 본 뒤 추가한다. 작은 학습 데이터에서는 optimizer가 전체 스캔을 고를 수도 있다.

요청 키는 대소문자를 구분하는 ascii_bin으로 둔다. 기본 문자열 collation에서 대소문자가 같다고 처리되어 멱등 키가 의도와 다르게 충돌하는 것을 피한다. 한글 제목은 utf8mb4다.

자료 제목과 URL은 카탈로그 서버에서 받아 **저장 당시 값의 사본**으로 남긴다. 목록 조회마다 다른 서버를 호출하지 않으며 그 서버의 DB에 FK를 걸지 않는다. 최신 제목 반영이 중요한 제품이라면 별도의 동기화 규칙이 필요하다.

## 트랜잭션

목록 INSERT만 성공하고 두 번째 항목 저장이 실패하면 일부만 저장된 목록이 생긴다. BeginTx 뒤 `tx.ExecContext`로 목록과 모든 항목을 넣고 끝에서 Commit한다. 실패하면 Rollback한다. 트랜잭션 안에서 `db.ExecContext`를 섞으면 별도 연결에서 실행될 수 있어 같은 원자적 작업이 아니다.

`defer tx.Rollback()`은 중간 오류의 정리를 맡는다. Commit 후의 Rollback은 이미 끝난 트랜잭션이므로 데이터에 추가 영향을 주지 않는다. 카탈로그 RPC는 트랜잭션 시작 전에 끝내 DB 연결과 잠금을 잡고 네트워크를 기다리지 않는다.

MySQL의 일반적인 기본 격리 수준은 REPEATABLE READ이지만 서버 설정을 확인해야 한다. 읽고 나서 쓰는 다단계 동작을 transaction으로 감쌌다고 모든 경쟁이 자동 해결되지는 않는다. 중복 생성은 애플리케이션 선조회보다 UNIQUE 제약으로 막는다. 행 잠금, 조건부 UPDATE와 재시도는 업무 규칙에 맞춰 선택한다.

## 같은 생성 요청이 두 번 오면

1. 정규화한 제목과 **자료 순서**로 fingerprint 계산
2. request_key를 UNIQUE 컬럼에 INSERT
3. 성공하면 항목 저장 후 Commit
4. 중복 키 오류 1062라면 현재 트랜잭션을 먼저 종료
5. 기존 요청의 fingerprint가 같으면 기존 목록 반환, 다르면 충돌

동시에 같은 키가 들어와도 DB의 UNIQUE 제약이 조정한다. HTTP 요청의 취소나 Commit 응답 유실은 ‘저장이 안 됐다’는 증거가 아니다. 같은 키로 재시도해 새 목록이 추가되지 않게 한다.

이 예제는 생성 효과를 중복하지 않는 형태다. 읽음 상태가 나중에 바뀌면 재요청 응답은 현재 목록 상태를 반환한다. 최초 응답 바이트를 영구 보존하는 완전한 응답 캐시는 아니다. 재요청도 카탈로그 확인을 먼저 하므로 카탈로그가 중단되면 재시도 응답 역시 실패할 수 있다. 기존 키 조회를 앞단에 추가하는 개선을 연습할 수 있다.

## UPDATE가 0행이면 없는 데이터인가

MySQL 드라이버의 기본 RowsAffected는 **실제 변경된 행 수**다. 이미 읽음인 항목을 다시 읽음으로 바꾸면 0일 수 있다. 이를 무조건 NotFound로 처리하면 반복 설정이 실패한다. 예제는 0일 때 해당 항목의 존재를 확인해 구분한다. `clientFoundRows=true` 같은 드라이버 설정으로 matched rows를 사용하는 대안도 있지만 설정의 의미를 알고 골라야 한다.

## 쿼리 살펴보기

전용 로컬 DB에서 실행한다.

```sql
SHOW INDEX FROM reading_items;

EXPLAIN SELECT resource_id, title, is_read
FROM reading_items
WHERE list_id = '실제_목록_ID'
ORDER BY position;

SELECT l.id, l.title, COUNT(*) AS item_count,
       SUM(i.is_read = TRUE) AS read_count
FROM reading_lists l
JOIN reading_items i ON i.list_id = l.id
GROUP BY l.id, l.title;
```

ORM과 sqlc는 SQL, 트랜잭션과 매핑을 이해한 뒤 비교한다. sqlc는 SQL에서 타입 있는 코드를 생성하는 선택지이고 ORM과 동일하지 않다. 이 작은 예제는 SQL이 실행되는 위치를 보려고 직접 적었다.

읽기: [Go DB 안내](https://go.dev/doc/database/), [트랜잭션](https://go.dev/doc/database/execute-transactions), [MySQL 드라이버](https://pkg.go.dev/github.com/go-sql-driver/mysql), [InnoDB 잠금](https://dev.mysql.com/doc/refman/8.4/en/innodb-locking.html).
