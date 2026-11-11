# 03. JPA 경험으로 GORM과 MySQL 읽기

이 장의 목표는 **Entity를 찾고, 목록 저장이 어디에서 트랜잭션으로 실행되는지 이해하는 것**입니다. 모든 명령은 go-practice 폴더에서 실행합니다.

## 1. 먼저 도구 이름을 역할에 연결해요

| 도구 | 역할 | Spring에서 떠올릴 것 |
| --- | --- | --- |
| GORM | 구조체와 DB 행을 매핑하고 SQL 실행 | ORM을 사용하는 목적 |
| MySQL 드라이버 | Go에서 MySQL과 통신 | JDBC 드라이버 |
| database/sql | DB 연결 풀과 공통 접근 기반 | 연결·DB 접근을 위한 기반 API |
| Goose | 버전 순서대로 스키마 SQL 적용 | Flyway나 Liquibase의 역할 |

GORM은 JPA와 같은 표준을 구현한 라이브러리는 아닙니다. ORM이라는 용도는 연결할 수 있지만 API와 객체 상태 관리 방식은 따로 익혀야 합니다.

## 2. Entity의 태그를 읽어 보세요

[internal/repository/mysql/model.go](../internal/repository/mysql/model.go)를 열면 다음과 같은 필드가 있습니다.

```go
ID    string `gorm:"column:id;primaryKey"`
Title string `gorm:"column:title"`
```

위는 구조체 필드 부분만 떼어 낸 것입니다. 백틱 안의 문자열은 Go의 **구조체 태그**입니다. GORM이 읽어 컬럼과 키 매핑에 사용합니다. Java의 @Column이나 @Id가 하던 역할을 떠올리면 됩니다.

파일에는 ListEntity와 ItemEntity가 있습니다. 자료 테이블 Entity는 [catalog_repository.go](../internal/repository/mysql/catalog_repository.go)의 ResourceEntity입니다.

- ListEntity: 목록의 ID, 제목, 요청 키, 생성 시간
- ItemEntity: 어느 목록의 어떤 자료인지, 항목 순서, 읽음 상태
- TableName(): 각각 reading_lists와 reading_items 테이블을 사용하도록 지정

업무 코드에서 사용하는 reading.List와 통신 DTO인 pb.ReadingList는 이 Entity와 별도입니다. 이름이 비슷하면 [1장의 타입 구분](01-spring-to-go.md)을 다시 보세요.

## 3. Repository 메서드를 읽는 순서

[internal/repository/mysql/reading_repository.go](../internal/repository/mysql/reading_repository.go)의 Get부터 읽으면 Create보다 짧습니다.

```go
r.db.WithContext(ctx).
    Preload("Items", func(db *gorm.DB) *gorm.DB {
        return db.Order("position")
    }).
    First(&entity, "id = ?", id)
```

실제 조회 체인을 줄을 나눠 표시했습니다.

| 부분 | 뜻 |
| --- | --- |
| WithContext(ctx) | 이번 요청의 취소·마감시간을 연결 |
| First(&entity, "id = ?", id) | ID 조건으로 한 목록을 읽고 entity에 채움 |
| Preload("Items", ...) | 목록의 항목도 명시적으로 조회 |
| Order("position") | 항목을 저장된 순서로 정렬 |

조건의 ?와 인자 id는 값을 바인딩하는 방식입니다. 사용자 값을 SQL 문자열에 직접 붙이는 방식으로 바꾸지 마세요.

Preload는 관련 항목을 별도 쿼리로 가져옵니다. Entity에 Items 필드가 있다는 사실만으로 모든 관계를 항상 자동 조회하는 것은 아닙니다. 조회 후에는 Entity를 업무 모델로 바꿔 반환합니다.

## 4. @Transactional에 해당하는 코드를 찾아요

Create에서 다음 패턴을 찾으세요.

```go
err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // 여기에서 목록과 항목을 tx로 저장합니다.
    return nil
})
```

위는 트랜잭션 범위를 보여 주는 축약 코드입니다. 실제 파일에는 두 종류의 INSERT가 들어 있습니다.

1. 목록 한 행을 저장합니다.
2. 선택한 자료의 항목들을 저장합니다.
3. 콜백이 nil을 반환하면 commit합니다.
4. 오류를 반환하면 rollback합니다.

트랜잭션 안에서는 받은 tx로 작업합니다. 바깥의 r.db를 사용하면 같은 트랜잭션에 참여하지 않는 작업이 생길 수 있습니다.

[UseCase.Create](../internal/reading/usecase.go)는 Catalog RPC를 먼저 끝낸 뒤 Repository를 호출합니다. 외부 서버 응답을 기다리는 동안 DB 트랜잭션을 잡고 있지 않도록 한 순서입니다.

## 5. Entity 필드만 바꾸면 UPDATE되나요?

이 코드에서는 저장 메서드를 명시적으로 호출해야 합니다. JPA의 영속 상태 Entity를 바꾸고 dirty checking으로 반영되기를 기대하던 흐름과 차이가 있습니다.

[SetRead](../internal/repository/mysql/reading_repository.go)는 Update("is_read", isRead)를 호출합니다. 이 표현은 false도 명시적으로 저장합니다. GORM의 Updates(struct)는 기본적으로 제로 값 필드를 생략할 수 있으므로 bool 수정에서 차이를 알아야 합니다.

또 MySQL은 실제로 변경된 행 수를 반환할 수 있습니다. 이미 true인 값을 다시 true로 설정하면 변경 행 수가 0이어도 항목이 존재할 수 있습니다. 그래서 이 Repository는 0행일 때 존재 여부를 확인합니다.

## 6. 같은 생성 요청을 두 번 보내면?

네트워크 응답을 받지 못했다고 해서 DB 저장도 실패한 것은 아닙니다. 재시도했을 때 목록이 두 개 생기지 않도록 requestKey를 사용합니다. 이런 중복 효과 방지를 **멱등성**이라고 부릅니다.

README의 같은 POST 요청을 그대로 다시 보내면 같은 ID를 돌려줍니다. 같은 키로 제목이나 자료 순서를 바꾸면 409를 응답합니다.

이 처리에는 세 가지가 협력합니다.

| 장치 | 역할 |
| --- | --- |
| request_key의 UNIQUE 제약 | 동시에 같은 키가 와도 중복 생성 방지 |
| fingerprint | 정규화한 제목과 자료 순서를 계산한 비교용 지문 |
| Repository.replay | 저장된 내용이 같으면 기존 목록 반환, 다르면 충돌 |

fingerprint의 해시 계산 방법까지 먼저 외울 필요는 없습니다. 같은 키에 같은 내용인지 비교하기 위한 값이라고 이해하세요. DB 제약은 [migrations/001_init.sql](../migrations/001_init.sql)에서 확인할 수 있습니다.

현재 재요청도 Catalog 조회를 먼저 합니다. Catalog 장애 중 재시도까지 처리하는 개선은 이후 과제입니다. 반환하는 것은 목록의 현재 상태이며 최초 응답 바이트를 그대로 보관한 캐시는 아닙니다.

## 7. 테이블은 언제 만들어지나요?

```bash
make db
make migrate
make migration-status
```

make db는 MySQL을 시작합니다. make migrate는 [별도의 명령](../cmd/migrate/main.go)을 실행해 Goose가 아직 적용하지 않은 SQL을 적용하게 합니다. make migration-status는 적용 상태를 보여 줍니다.

서버 시작 때 GORM AutoMigrate를 실행하는 구성은 아닙니다. Entity 정의와 스키마 변경 이력을 나눠 관리합니다. Compose로 전체 실행하면 migrate 컨테이너가 성공한 뒤 Catalog와 Reading이 시작됩니다. [002_catalog.sql](../migrations/002_catalog.sql)은 자료 테이블만 생성하며 초기 자료를 넣지 않습니다.

DB 구조를 바꿀 때는 다음 번호의 migration 파일을 추가합니다. 기존에 적용한 파일을 계속 덮어쓰는 방식으로 운영하지 않습니다.

## 8. 연결 풀은 어느 파일에 있나요?

[internal/repository/mysql/database.go](../internal/repository/mysql/database.go)가 GORM과 DB 연결 풀을 만듭니다. 기본 최대 연결은 10개, 유지할 유휴 연결은 5개입니다. 서버가 시작될 때 연결을 확인하고 종료할 때 닫습니다.

goroutine을 많이 시작한다고 DB 연결 제한이 사라지는 것은 아닙니다. 연결이 모두 사용 중이면 반환될 때까지 기다리는 작업이 생길 수 있습니다. 이 설정은 [환경 설정 장](04-architecture.md)에서 이어집니다.

## 9. 자료 등록과 목록의 복사본

[CatalogRepository.Upsert](../internal/repository/mysql/catalog_repository.go)는 GORM의 OnConflict로 같은 ID의 제목과 URL을 갱신합니다. PUT 요청을 반복해도 같은 ID의 자료가 여러 행 생기지 않습니다. Batch는 DB의 IN 조회 결과를 입력 ID 순서로 다시 정렬합니다.

Reading은 목록 생성 시 Catalog에서 받은 제목과 URL을 항목에 복사합니다. 따라서 Catalog 자료를 수정해도 기존 목록은 생성 당시 정보를 유지하고, Catalog가 중단돼도 기존 목록 조회는 가능합니다. 같은 로컬 DB를 쓰더라도 두 앱의 테이블 사이 FK나 직접 JOIN은 두지 않았습니다.

## 10. 쿼리가 복잡해지면 어떻게 하나요?

이 프로젝트는 Spring 개발자가 Entity와 Repository를 연결하기 쉽게 GORM을 선택했습니다. 기본 CRUD는 메서드 체인으로 작성하고, 복잡한 조회는 별도 Repository 메서드에 둡니다.

Go에서 ORM이 필수는 아닙니다. [sqlc](https://docs.sqlc.dev/en/stable/)는 직접 작성한 SQL에서 타입이 있는 Go 코드를 생성하고, [Ent](https://entgo.io/docs/getting-started/)는 스키마를 코드로 정의해 ORM 코드를 생성합니다. SQL 중심 팀인지, 관계·모델 중심 팀인지에 따라 비교할 수 있습니다. 이들은 선택지이며 현재 프로젝트에 모두 함께 설치한 것은 아닙니다.

## 확인 문제

1. 목록은 저장됐지만 항목 저장이 실패하면 왜 rollback이 필요한가요?
2. isRead를 false로 바꿀 때 Update를 명시적으로 쓰는 이유는 무엇인가요?
3. 같은 요청이 동시에 들어오는 문제를 Go의 if 검사만으로 해결하지 않는 이유는 무엇인가요?

답: 목록과 항목을 한 작업으로 저장해야 하기 때문입니다. false는 생략되기 쉬운 제로 값이고, 중복 요청은 DB의 UNIQUE 제약이 동시 상황까지 조정하도록 하기 때문입니다.

[이전: HTTP와 gRPC](02-grpc.md) · [다음: DI·설정·실행](04-architecture.md)

더 읽기: [GORM 트랜잭션](https://gorm.io/docs/transactions.html), [GORM 업데이트](https://gorm.io/docs/update.html), [Goose](https://github.com/pressly/goose).
