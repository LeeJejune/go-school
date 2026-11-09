# 05. Spring 개발자를 위한 실무 코드 안내

## 어떤 도구가 무엇을 하는가

이 예제의 실행 스택은 **Kratos v3 + Fx + GORM + Goose**다. Kratos는 transport와 미들웨어, Fx는 DI와 프로세스 생명주기를 담당한다. Kratos App과 Fx가 각각 신호를 받게 중복 구성하지 않았다. ORM과 마이그레이션은 별도 책임이다.

| Spring에서 익숙한 것 | 현재 코드 |
| --- | --- |
| SpringApplication.run | cmd/*/main.go의 fx.New(...).Run() |
| @Configuration, @Bean, 생성자 주입 | internal/app/app.go의 fx.Provide, fx.As |
| application.yml / @ConfigurationProperties | internal/config/config.go의 타입과 env 태그 |
| Controller | internal/service/reading.go |
| 업무 Service | internal/reading/usecase.go |
| Domain model | internal/reading/model.go |
| Entity | internal/data/model.go |
| Repository 구현 | internal/data/reading_repository.go |
| @Transactional | GORM Transaction 콜백 |
| Flyway / Liquibase | Goose, cmd/migrate와 migrations |
| Feign 같은 외부 클라이언트 | internal/catalogclient/client.go |
| 공통 interceptor / exception handler | Kratos middleware, service의 오류 매핑 |
| CompletableFuture로 독립 호출 결합 | gateway/dashboard.go의 errgroup |

## 처음 열 파일

**internal/app/app.go → internal/service/reading.go → internal/reading/usecase.go → internal/data/reading_repository.go** 순서로 읽는다.

main.go는 시작 명령일 뿐이다. 의존성이 어떻게 연결되는지는 app 모듈에서, 실제 동작은 아래의 요청 경로에서 찾는다.

## 목록 생성 요청을 추적한다

1. api/http.yaml의 POST /v1/lists가 CreateList RPC로 매핑된다.
2. Gateway가 생성된 HandlerClient로 실제 Reading RPC를 호출한다.
3. service.Reading.CreateList가 Protobuf 요청을 reading.CreateInput으로 바꾼다.
4. reading.UseCase.Create가 제목, 요청 키, 자료 ID를 검증한다.
5. reading.Catalog 인터페이스를 통해 Catalog RPC를 호출한다.
6. 응답의 자료 정보를 이용해 fingerprint와 CreateParams를 만든다.
7. reading.Repository 인터페이스의 Create를 호출한다.
8. data.ReadingRepository가 GORM Entity를 만들고 트랜잭션으로 저장한다.
9. Entity → reading.List → pb.ReadingList 순서로 바뀌어 JSON 응답으로 돌아온다.

## model.go와 proto가 비슷하게 보이는 이유

- school.proto: 공개 요청·응답과 RPC 계약. 생성 결과는 gen이다.
- reading/model.go: 업무 모델과 필요한 인터페이스.
- data/model.go: 테이블과 매핑할 ORM Entity.
- migrations/*.sql: DB 스키마와 인덱스·제약의 변경 이력.

동일한 자료를 경계마다 다른 목적의 타입으로 표현한다. 통신의 시간은 string, 업무·DB 시간은 time.Time이라는 차이도 있다. 모델이 자동으로 함께 변경되는 구조는 아니다.

## DI를 실제로 읽는다

app.Reading 안의 다음 선언은 GORM 구현체를 업무 코드의 Repository로 주입하라는 뜻이다.

```go
fx.Annotate(data.NewReadingRepository, fx.As(new(reading.Repository)))
```

UseCase의 생성자는 Repository와 Catalog를 요구한다. Fx가 등록된 생성자들을 연결한다. DB와 클라이언트의 OnStop 훅은 서버 종료 뒤 실행된다. 단위 테스트에서는 이 컨테이너 없이 대역을 생성자에 넣는다.

## ORM도 SQL과 트랜잭션 이해가 필요하다

data/model.go에서 테이블명을, Repository에서 Preload·정렬·Create·Update를 확인한다. GORM의 Transaction 콜백은 오류 시 rollback한다. true를 false로 바꾸는 업데이트도 명시적으로 포함한다.

AutoMigrate는 호출하지 않는다. make migrate가 Goose 이력을 확인하고 새 스키마 변경을 적용한다. Compose에서도 migrate 서비스가 완료되어야 Reading이 시작된다.

## 서비스에서 goroutine을 직접 사용하는 예

GET /v1/dashboard/{id}는 새로 추가한 Gateway의 조합 API다.

```text
                    ┌─ Catalog.ListResources ──┐
HTTP → errgroup ────┤                         ├─ Wait → JSON
                    └─ Reading.GetList ────────┘
```

둘은 서로의 결과가 필요 없으므로 동시에 실행한다. 한쪽이 실패하면 다른 쪽에 취소 신호를 전달하고, 둘 다 종료된 뒤 오류를 반환한다. HTTP 요청 취소도 각 RPC에 전달된다.

이 코드는 응답 전에 필요한 작업을 동시에 처리하는 예다. 프로세스가 종료되어도 반드시 보장되어야 하는 알림·결제 같은 작업의 영속 큐를 대체하지 않는다. 대량 작업은 DB/외부 API의 한계를 고려해 동시 실행 수를 제한한다. 별도의 worker pool 예제는 [go-basic](../../go-basic/README.md)에 있다.

## 개발 작업을 추가하는 순서

새 RPC 기능은 proto → 코드 생성 → Controller → UseCase → Repository 순서로 연결하고 app에 필요한 provider를 등록한다. DB 변경은 새 Goose migration으로 추가한다. 업무 검증은 단위 테스트, ORM 동작은 MySQL 통합 테스트, 외부 계약은 API 통합 테스트로 확인한다.

실행에는 go-practice 폴더에서 make db → make migrate → 터미널별 make run-catalog / run-reading / run-gateway를 사용한다. API 예시는 [README](../README.md), 검증 명령은 [테스트 안내](04-testing.md)를 본다.

도구 공식 문서: [Kratos](https://go-kratos.dev/docs/), [Fx](https://uber-go.github.io/fx/), [GORM](https://gorm.io/docs/), [Goose](https://github.com/pressly/goose), [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup).
