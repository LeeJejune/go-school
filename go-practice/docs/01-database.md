# 01. GORM, MySQL과 마이그레이션

## 코드에서 먼저 볼 곳

1. `internal/data/model.go`: DB 테이블과 매핑하는 GORM Entity
2. `internal/data/reading_repository.go`: ORM 조회·저장과 트랜잭션
3. `internal/data/database.go`: 연결 풀과 Fx 시작·종료 훅
4. `migrations/001_init.sql`: 실제 제약과 인덱스
5. `cmd/migrate/main.go`: Goose 실행 명령

## GORM과 database/sql

GORM은 SQL 생성과 결과 매핑을 담당하는 ORM이다. 아래에는 여전히 MySQL 드라이버와 `database/sql` 연결 풀이 있다. `*gorm.DB`는 애플리케이션 전체에서 공유하고, 요청마다 `WithContext(ctx)`로 요청 범위를 만든다.

Spring의 JPA처럼 Entity 필드를 바꾸기만 하면 저장되는 구조는 아니다. 변경은 `Create`, `Update` 같은 메서드로 명시한다. 테이블 구조는 SQL 마이그레이션이 관리한다.

`NewDB`는 Fx에 DB를 제공한다. OnStart에서 Ping으로 연결을 확인하고 OnStop에서 풀을 닫는다. 풀은 열린 연결 10개, 유휴 연결 5개가 기본이며 환경 변수로 조절한다. DSN의 `parseTime=true`는 시간 필드를 time.Time으로 읽기 위한 옵션이다.

## Entity, 업무 모델, API DTO

`ListEntity`와 `ItemEntity`는 테이블 매핑용이다. `reading.List`는 업무 코드가 사용하고, `pb.ReadingList`는 통신에 사용한다. 각각 DB·업무·통신의 변경 이유를 갖는다.

`Preload("Items")`는 목록과 항목 관계를 읽는다. 항목에는 `position` 정렬을 명시한다. Preload는 별도 쿼리를 실행하므로 ORM 메서드 수와 SQL 실행 횟수가 항상 같지는 않다.

## ORM 트랜잭션

```go
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&list).Error; err != nil {
        return err
    }
    return tx.Create(&items).Error
})
```

콜백이 nil을 반환하면 commit, 오류를 반환하면 rollback한다. 트랜잭션 안에서는 전달받은 tx를 사용한다. 카탈로그 RPC는 UseCase에서 먼저 끝내므로 DB 연결과 잠금을 잡고 외부 서버를 기다리지 않는다.

목록 키가 중복될 때는 트랜잭션 종료 후 기존 데이터를 조회한다. 항목의 중복 키 오류를 목록 생성 재시도로 오인하지 않도록 구분한다.

## 중복 요청과 스키마 제약

- request_key UNIQUE: 같은 생성 요청의 중복 효과 방지
- fingerprint: 정규화한 제목과 자료 순서로 계산
- 항목 PK (list_id, resource_id): 같은 자료의 중복 등록 방지
- UNIQUE (list_id, position): 항목 순서 중복 방지
- FK와 ON DELETE CASCADE: 목록 삭제 시 항목 정리

요청 키와 ID는 ascii_bin으로 비교하고 제목은 utf8mb4로 저장한다. 자료의 제목과 URL은 저장 시점의 사본이다.

같은 요청 키라도 다른 fingerprint면 충돌이다. 이 처리는 애플리케이션의 선조회만으로 보호하지 않고 MySQL의 UNIQUE 제약을 이용한다. Commit 응답 유실이 저장 실패를 뜻하지 않기 때문에 재시도에는 같은 키를 사용한다.

재시도 응답은 저장된 목록의 현재 상태다. 원래 응답 바이트의 캐시는 아니다. 현재는 재시도도 Catalog를 먼저 확인한다. Catalog 장애 중에도 재시도가 가능하도록 기존 키 확인 순서를 개선하는 과제를 남겨 두었다.

## false 업데이트와 RowsAffected

GORM의 `Updates(struct)`는 기본적으로 제로 값을 생략할 수 있다. 이 프로젝트는 `Update("is_read", isRead)`로 false도 명시적으로 저장한다.

MySQL은 실제 변경한 행 수를 반환하므로 이미 true인 항목을 true로 바꾸면 0행일 수 있다. 0행일 때 존재 여부를 조회해 없는 항목과 반복 요청을 구분한다.

## Goose 마이그레이션

```bash
make db
make migrate
make migration-status
```

Goose는 버전 이력 테이블을 관리하고 적용하지 않은 SQL만 실행한다. SQL은 바이너리에 embed되어 작업 디렉터리에 의존하지 않는다. MySQL DDL의 특성 때문에 초기 마이그레이션은 NO TRANSACTION이며, MySQL 테이블 잠금으로 동시 마이그레이션 실행을 조정한다.

첫 파일의 IF NOT EXISTS는 기존 학습 예제 스키마와 새 DB를 모두 지원하기 위한 baseline이다. 임의의 스키마 차이를 자동 교정하지는 않는다. 다음 변경은 002번 파일로 추가한다. 서버 시작 시 AutoMigrate를 실행하지 않고 배포의 일회성 단계에서 migrate를 실행한다.

## SQL도 확인한다

로컬 학습 DB에서 아래를 실행해 ORM이 사용하는 인덱스와 조회 계획을 본다.

```sql
SHOW INDEX FROM reading_items;
EXPLAIN SELECT * FROM reading_items
WHERE list_id = '실제_목록_ID' ORDER BY position;
```

ORM이 SQL 이해를 대체하지 않는다. 복잡한 조회는 GORM의 명시적 SQL이나 sqlc를 비교할 수 있다. 현재 저장소는 ORM 사용 흐름을 배우기 위해 GORM을 선택했다.

공식 문서: [GORM](https://gorm.io/docs/), [트랜잭션](https://gorm.io/docs/transactions.html), [Goose](https://github.com/pressly/goose), [Go DB](https://go.dev/doc/database/).
