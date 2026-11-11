# go-practice — Spring 개발자의 첫 Go 백엔드

Java/Spring 개발자가 **HTTP Controller → 업무 Service → Repository → DB** 흐름을 Go에서 직접 따라가는 실습입니다. HTTP는 Gin, 서버 간 통신은 grpc-go, DI는 Fx, ORM은 GORM을 사용합니다. 문법이 낯설면 [go-basic 1~4장](../go-basic/README.md)부터 읽으세요.

## 1. 서버 실행

Go 1.26 이상과 Docker Desktop/Compose를 준비합니다. Docker Desktop을 켜고 저장소 루트에서 실행합니다.

```bash
cd go-practice
make up
```

MySQL → Goose migration → Catalog/Reading/Gateway 순서로 시작합니다. 이후 명령은 모두 go-practice 폴더 기준입니다.

```bash
curl -sS http://localhost:8080/v1/resources
```

처음에는 `{"resources":[]}`가 정상입니다. **서버가 자료를 자동으로 넣지 않습니다.** 이전에 등록한 자료가 있다면 DB에 보관된 값이 나옵니다.

## 2. HTTP 파일이나 Postman으로 실습

서버 코드를 수정하지 않고 아래 중 하나로 요청을 보내세요. 등록할 제목과 URL도 요청 본문에서 바꿉니다.

| 도구 | 열 파일 | 사용법 |
| --- | --- | --- |
| GoLand / IntelliJ HTTP Client | [requests/school.http](requests/school.http) | 각 요청 왼쪽 실행 버튼으로 위에서 아래 순서대로 실행 |
| VS Code REST Client 확장 | [requests/school.http](requests/school.http) | 각 요청의 Send Request 클릭 |
| Postman | [컬렉션 JSON](requests/go-practice.postman_collection.json) | Import → 파일 선택 → 컬렉션 순서대로 Send 또는 Runner 실행 |

요청 순서는 **자료 두 개 등록 → 자료 조회 → 목록 생성 → 목록 조회 → 읽음 변경 → 대시보드 → 미읽음 변경 → 조회**입니다.

- HTTP 파일: 목록 생성 응답의 `readingList.id`를 파일 위쪽 `@listID`에 복사합니다.
- Postman: 목록 생성 요청의 응답 스크립트가 컬렉션 변수 `listID`를 자동 저장합니다.
- 새 목록을 만들려면 `requestKey`를 바꿉니다. 같은 키와 같은 내용은 같은 목록, 같은 키에 다른 내용은 409입니다.
- 자료 ID가 같으면 PUT은 제목과 URL을 갱신합니다. 기존 읽기 목록의 자료 정보는 생성 시점의 복사본을 유지합니다.

| HTTP API | 하는 일 |
| --- | --- |
| PUT /v1/resources/{id} | 자료 등록 또는 수정 |
| GET /v1/resources | 등록된 자료 조회 |
| POST /v1/lists | 읽기 목록 생성 |
| GET /v1/lists/{id} | 목록 조회 |
| PATCH /v1/lists/{id}/items/{resourceID} | 읽음 상태 지정 |
| GET /v1/dashboard/{id} | 카탈로그와 목록을 동시 조회해 조합 |

PATCH 성공 응답은 `{}`입니다. `isRead`를 true/false로 보내고 GET으로 확인하세요. 이 계약에서는 생략한 `isRead`도 false로 해석합니다.

## 3. Spring에서 찾던 코드는 여기 있습니다

| 익숙한 역할 | 처음 열 파일 |
| --- | --- |
| @RestController / URL 매핑 | [controller/http/router.go](internal/controller/http/router.go), [reading.go](internal/controller/http/reading.go) |
| HTTP DTO | [controller/http/dto.go](internal/controller/http/dto.go) |
| 서버 간 RPC Controller | [controller/grpc/reading.go](internal/controller/grpc/reading.go) |
| @Service | [reading/usecase.go](internal/reading/usecase.go), [catalog/catalog.go](internal/catalog/catalog.go) |
| Repository 인터페이스 | [reading/model.go](internal/reading/model.go) |
| Repository 구현 / Entity | [repository/mysql](internal/repository/mysql) |
| RPC DTO 계약 | [school.proto](api/proto/school/v1/school.proto) |
| @Configuration / @Bean | [app/app.go](internal/app/app.go) |

HTTP와 gRPC Controller는 입력·출력을 바꾸고, 업무 규칙은 reading/catalog에 둡니다. 기능이 늘면 업무 패키지를 추가합니다. 구조를 고른 근거는 [국내외 프로젝트 비교](docs/07-project-comparison.md)에 있습니다.

## 4. 학습 순서

| 문서 | 답할 질문 |
| --- | --- |
| [01 Spring 역할 대응](docs/01-spring-to-go.md) | Controller, DTO, Entity, pb, *와 &는 무엇인가요? |
| [02 HTTP·gRPC·goroutine](docs/02-grpc.md) | 요청이 서버들을 어떻게 거치나요? |
| [03 GORM·MySQL](docs/03-database.md) | 쿼리, 트랜잭션, 스키마는 어디서 관리하나요? |
| [04 DI·설정·생명주기](docs/04-architecture.md) | main.go 뒤에서 누가 객체와 서버를 준비하나요? |
| [05 테스트](docs/05-testing.md) | 단위 테스트와 실제 DB 검사는 어떻게 하나요? |
| [06 공식 자료](docs/06-reading.md) | 막힐 때 무엇을 읽나요? |
| [07 국내외 구조 비교](docs/07-project-comparison.md) | 어떤 공개 코드를 참고했고 왜 이 구조를 골랐나요? |

## 5. Go 서버를 직접 디버깅하기

전체 Docker 실행을 사용했다면 먼저 `make down`으로 같은 포트를 비웁니다.

```bash
make deps
make db
make migrate
make migration-status
```

터미널 세 개를 go-practice 폴더로 이동한 뒤 각각 실행합니다.

```bash
# 터미널 1
make run-catalog
```

```bash
# 터미널 2
make run-reading
```

```bash
# 터미널 3
make run-gateway
```

요청을 기다리며 종료되지 않는 것이 정상입니다. `.vscode/launch.json`의 서버별 설정으로도 디버깅할 수 있습니다. 목록 생성에 중단점을 걸고 HTTP 파일의 요청을 보내 보세요.

## 6. 도구와 검사 명령

| 도구 | 역할 |
| --- | --- |
| Gin | HTTP 라우팅, JSON 바인딩·검증, 미들웨어 |
| grpc-go + go-grpc-middleware | RPC 서버·클라이언트, 로그·복구·제한 시간 |
| Fx | 생성자 DI와 시작·종료 관리 |
| GORM + MySQL | Entity 매핑, DB 접근·트랜잭션 |
| Goose | SQL migration |
| caarlos0/env | 환경 변수 설정 바인딩 |
| Protobuf + Buf | RPC 계약과 Go 코드 생성 |
| errgroup | 독립 조회의 동시 실행·취소·완료 대기 |

버전은 [go.mod](go.mod)에 고정합니다. Gin은 HTTP 계층을 담당하므로 gRPC 서버에는 별도 grpc-go를 사용합니다. 사용자 인증·소유권·TLS는 이후 확장 과제입니다. `/healthz`는 프로세스 응답만 확인합니다.

```bash
make test
make check

# 실제 MySQL 및 전체 HTTP→gRPC→DB 검사
make db
make migrate
make test-db
```

`.proto`를 바꿀 때만 `make tools` → `make generate`를 실행합니다. 생성 결과인 gen은 직접 고치지 않습니다. HTTP 경로·DTO는 controller/http에서 수정합니다.

```bash
# 로그 보기: Ctrl+C로 로그 보기 종료
docker compose logs -f catalog reading gateway

# 컨테이너 종료, DB 볼륨 유지
make down
```

기본 설정으로 바로 실행할 수 있고 `.env`는 자동 로딩하지 않습니다. [환경 설정](docs/04-architecture.md)을 참고하세요.

[저장소 안내](../README.md) · [기초 Go](../go-basic/README.md) · [커밋 규칙](../CONTRIBUTING.md)
