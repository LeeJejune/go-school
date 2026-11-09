# go-school

Go 문법부터 동시성, MySQL, gRPC와 grpc-gateway까지 이어지는 백엔드 학습 저장소.

예제는 **읽을거리 목록**이다. 카탈로그에서 자료를 고르고, 목록을 만들고, 읽음 상태를 바꾼다. 기능은 작게 두고 요청이 서버와 DB를 거치는 과정에 집중한다.

## 학습 순서

| 순서 | 내용 | 문서 | 코드 |
| --- | --- | --- | --- |
| 1 | 변수, slice와 map, 포인터, 메서드, interface, error | [Go 기본 문법](docs/01-go-basics.md) | [기본 예제](examples/01-basics/main.go) |
| 2 | goroutine, channel, Mutex, Context, worker pool | [동시성과 취소](docs/02-concurrency.md) | [worker pool](internal/workpool/map.go), [사용 예제](examples/02-concurrency/main.go) |
| 3 | database/sql, 커넥션 풀, SQL, 트랜잭션, 멱등성 | [MySQL과 데이터 접근](docs/03-mysql.md) | [SQL 저장소](internal/mysqlstore/store.go), [스키마](migrations/001_init.sql) |
| 4 | Protobuf, gRPC, Gateway, deadline과 오류 | [서버 간 통신](docs/04-grpc.md) | [계약](api/proto/school/v1/school.proto), [클라이언트](internal/catalogclient/client.go) |
| 5 | 패키지 경계, 의존성 방향, 클린 아키텍처의 필요성 | [아키텍처 선택](docs/05-architecture.md) | [목록 서비스](internal/reading/service.go) |
| 6 | 테스트, 종료 처리, 학습 과제 | [테스트와 다음 단계](docs/06-testing.md) | [테스트 코드](internal/reading/service_test.go) |

## 구성

```text
curl ── HTTP/JSON ── Gateway :8080
                        ├── gRPC ── Catalog :9001
                        └── gRPC ── Reading :9002 ── MySQL :3307
                                        └── gRPC ── Catalog
```

- **Catalog**: 공식 문서 4개를 제공하는 읽기 전용 카탈로그. 자료는 코드에 고정
- **Reading**: 자료를 확인한 뒤 목록과 항목을 MySQL 트랜잭션으로 저장. 요청 키로 재전송 처리
- **Gateway**: 생성된 클라이언트로 실제 gRPC 호출. HTTP 핸들러에서 서비스 함수를 바로 부르지 않음

서버 2개와 Gateway는 독립 프로세스이며 Compose에서는 각각 별도 컨테이너다. 서버 간 통신을 연습하기 위한 분리다. 이 크기의 실제 제품이라면 한 프로세스에서 시작하는 편이 단순하다.

## 준비

- Go **1.26.0 이상**. 의존성 버전은 `go.mod`, `go.sum`에 고정
- MySQL **8.4**. 아래 실행 예시는 Docker Compose 기준
- curl. 브라우저나 프론트엔드는 필요 없음
- 코드 생성 파일을 포함하므로 실행만 할 때 Buf와 protoc 설치는 불필요

IDE는 **GoLand** 또는 **VS Code + Go 확장**. JetBrains 환경이 익숙하면 GoLand가 편하고, 무료 환경은 VS Code로 충분하다. 학생용 GoLand 라이선스는 자격 확인이 필요하다. [공식 설정과 읽을거리](docs/07-reading.md)

## 실행

전체 컨테이너:

```bash
docker compose up --build -d
docker compose logs -f reading gateway
```

디버깅할 때는 MySQL만 컨테이너로 두고 서버를 터미널 3개에서 각각 실행한다.

```bash
make db
go run ./cmd/catalog
```

```bash
go run ./cmd/reading
```

```bash
go run ./cmd/gateway
```

기본 DB는 `school:school@127.0.0.1:3307/school`이다. 로컬 학습용 계정이며 Compose 포트는 localhost에만 공개한다. 다른 DB를 쓸 때는 [환경 변수](docs/04-grpc.md#환경-변수)를 참고한다.

## API 사용

```bash
# 자료 목록
curl -sS http://localhost:8080/v1/resources

# 읽을거리 목록 생성
curl -sS -X POST http://localhost:8080/v1/lists \
  -H 'Content-Type: application/json' \
  -d '{"requestKey":"go-study-001","title":"이번 주 Go 공부","resourceIds":["go-tour","go-context"]}'
```

응답의 `readingList.id`를 다음 URL에 넣는다.

```bash
curl -sS http://localhost:8080/v1/lists/LIST_ID

curl -sS -X PATCH http://localhost:8080/v1/lists/LIST_ID/items/go-tour \
  -H 'Content-Type: application/json' \
  -d '{"isRead":true}'
```

같은 `requestKey`와 같은 제목/자료 순서로 다시 요청하면 기존 목록을 반환한다. 같은 키에 다른 내용을 보내면 HTTP 409다. 읽음 처리는 같은 값을 반복 설정해도 성공한다.

## 자주 쓰는 명령

```bash
make basics       # 기본 문법 예제
make concurrency  # worker pool 예제
make build        # 서버 바이너리 생성
make check        # vet + race detector를 켠 테스트
make test-db      # 전용 로컬 MySQL로 SQL 통합 테스트
make down         # 컨테이너 종료, DB 볼륨 유지
```

프로토콜을 수정할 때만 생성 도구를 설치한다.

```bash
make tools
make generate
```

Buf는 설치된 Go 버전과 맞는 1.50.0으로 고정했다. 생성기와 Gateway 런타임 버전도 고정한다. 생성한 `gen/`은 직접 수정하지 않는다.

## 코드 읽는 순서

1. `school.proto`에서 요청과 응답 확인
2. `cmd/reading/main.go`에서 DB, 클라이언트와 서비스 연결 확인
3. `reading/service.go`에서 검증 → 카탈로그 조회 → 저장 흐름 확인
4. `mysqlstore/store.go`에서 트랜잭션과 SQL 확인
5. 테스트에서 실패, 중복, 취소의 기대 동작 확인

## 범위와 검증 상태

컴파일과 정적 분석, Protobuf 검사, Compose 설정 파싱은 통과했다. 서버 기동, 테스트 실행과 MySQL 연동은 아직 확인하지 않았다. 상세 결과는 [검증 상태](VALIDATION.md)에 남겼다.

인증, TLS, 사용자별 목록 소유권, 운영용 마이그레이션 도구는 포함하지 않는다. 로컬 학습용이며 외부에 그대로 배포하지 않는다. DB 초기 SQL은 **빈 볼륨의 최초 기동**에만 적용된다. 기존 데이터가 있는 DB에는 자동 재적용되지 않는다.
