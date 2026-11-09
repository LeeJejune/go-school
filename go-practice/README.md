# go-practice

실무형 Go 백엔드의 구성과 개발 흐름을 연습하는 독립 프로젝트다. 예제는 공부할 자료를 골라 읽을거리 목록을 만들고 읽음 상태를 관리하는 서비스다.

서버 프레임워크, DI, ORM, 마이그레이션 도구를 사용하는 하나의 참조 구성이다. 팀마다 다른 조합을 선택할 수 있다.

## 적용한 도구

| 영역 | 도구 | 실제 적용 위치 |
| --- | --- | --- |
| HTTP/gRPC 서버, 제한 시간, 복구·로깅 미들웨어 | Kratos v3 | [internal/server](internal/server) |
| DI 컨테이너, 모듈, 시작·종료 순서 | Uber Fx | [internal/app/app.go](internal/app/app.go) |
| ORM, 관계 조회, 트랜잭션 | GORM + MySQL | [internal/data](internal/data) |
| 버전별 DB 스키마 변경 | Goose | [migrations](migrations), [cmd/migrate](cmd/migrate) |
| 타입이 있는 환경 설정 | caarlos0/env | [internal/config/config.go](internal/config/config.go) |
| RPC 계약과 HTTP/JSON 변환 | Protobuf, grpc-go, grpc-gateway | [api](api), [gen](gen) |
| 독립적인 RPC의 동시 실행·취소 | errgroup | [internal/gateway/dashboard.go](internal/gateway/dashboard.go) |

의존성 버전은 [go.mod](go.mod)와 go.sum에 고정한다. **처음에는 [Spring 개발자를 위한 실무 코드 안내](docs/05-spring-guide.md)를 읽는다.**

## 구성

```text
HTTP/JSON → Gateway :8080
              ├─ gRPC → Catalog :9001  (자료 카탈로그)
              └─ gRPC → Reading :9002  (공부 목록)
                            ├─ gRPC → Catalog
                            └─ GORM → MySQL :3307

Goose migration 명령 → MySQL
```

Catalog, Reading, Gateway는 별도 애플리케이션이다. Controller/Service/Repository에 해당하는 역할은 각 애플리케이션 내부에서 나뉜다.

- **Catalog**: 코드에 정의한 공식 자료 4개를 조회한다.
- **Reading**: 입력 검증, 자료 확인, 목록 저장과 읽음 상태 변경을 담당한다.
- **Gateway**: 기존 HTTP API를 실제 gRPC 호출에 연결한다. 대시보드 API에서는 두 서버의 응답을 동시에 조회해 합친다.
- **Migrate**: 서버와 별도로 실행하는 일회성 스키마 변경 명령이다. Compose에서는 성공 후 Reading이 시작된다.

## 코드 읽는 순서

| Spring에서 익숙한 역할 | 볼 코드 |
| --- | --- |
| 부팅과 DI 설정 | [internal/app/app.go](internal/app/app.go) |
| Controller | [internal/service/reading.go](internal/service/reading.go) |
| 업무 Service | [internal/reading/usecase.go](internal/reading/usecase.go) |
| 업무 모델·Repository 계약 | [internal/reading/model.go](internal/reading/model.go) |
| ORM Entity | [internal/data/model.go](internal/data/model.go) |
| Repository 구현 | [internal/data/reading_repository.go](internal/data/reading_repository.go) |
| API 요청·응답 DTO 원본 | [school.proto](api/proto/school/v1/school.proto) |
| 외부 서비스 클라이언트 | [internal/catalogclient/client.go](internal/catalogclient/client.go) |
| 동시 호출을 하는 BFF 기능 | [internal/gateway/dashboard.go](internal/gateway/dashboard.go) |

`cmd/*/main.go`는 Fx 모듈을 실행하는 짧은 진입점이다. 업무 코드는 DI 컨테이너를 직접 조회하지 않고 생성자로 전달받는다.

## 준비

- Go **1.26.0 이상**, Docker Desktop/Compose, curl
- IDE에서 go-practice 폴더를 연다. VS Code는 Go 확장과 서버별 디버그 구성을 제공한다.
- 실행에는 Buf/protoc가 필요 없다. 생성한 코드가 저장소에 포함되어 있다.
- 기본 환경 변수로 바로 실행된다. 설정을 바꾸려면 [.env.example](.env.example)과 [환경 변수](docs/02-grpc.md#환경-변수)를 본다. `.env`는 자동으로 읽지 않는다.

## 로컬 개발 실행

먼저 Docker Desktop을 켜고 go-practice 폴더에서 DB와 마이그레이션을 준비한다. 아래 명령은 모두 이 폴더에서 실행한다.

```bash
# 저장소 루트라면 먼저 cd go-practice
make deps
make db
make migrate
make migration-status
```

터미널 세 개에서 각각 실행한다.

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

DB는 `127.0.0.1:3307`, 로컬 학습 계정은 `school / school`, DB 이름은 `school`이다.
서버는 `Ctrl+C`로 종료한다. Fx가 HTTP/gRPC 서버 → 클라이언트·DB 순서로 정리한다.

모든 애플리케이션을 컨테이너에서 실행하려면 다음 명령을 사용한다. 로컬 서버와 같은 포트를 사용하므로 로컬 서버를 먼저 종료한다.

```bash
make up
docker compose logs -f catalog reading gateway
make down
```

`make down`은 DB 볼륨을 유지한다. Compose 프로젝트 이름 go-school도 유지해 이전 학습 DB를 그대로 사용한다. 기존 예제 스키마에도 첫 Goose 마이그레이션을 적용할 수 있다. 이후 스키마 변경은 새 번호의 마이그레이션 파일로 추가한다. GORM AutoMigrate는 실행하지 않는다.

## API 실습

```bash
curl -sS http://localhost:8080/v1/resources

curl -sS -X POST http://localhost:8080/v1/lists \
  -H 'Content-Type: application/json' \
  -d '{"requestKey":"go-study-001","title":"이번 주 Go 공부","resourceIds":["go-tour","go-context"]}'
```

응답의 `readingList.id`를 아래 `LIST_ID` 대신 넣는다.

```bash
curl -sS http://localhost:8080/v1/lists/LIST_ID

curl -sS -X PATCH http://localhost:8080/v1/lists/LIST_ID/items/go-tour \
  -H 'Content-Type: application/json' \
  -d '{"isRead":true}'

# Catalog와 Reading을 errgroup으로 동시에 호출하는 API
curl -sS http://localhost:8080/v1/dashboard/LIST_ID
```

같은 요청 키와 같은 제목·자료 순서로 생성 요청을 반복하면 같은 목록을 반환한다. 같은 키에 다른 내용을 보내면 409다. `isRead`는 true/false를 반복 설정해도 성공한다.

## 테스트와 개발 명령

```bash
make check        # go vet + race detector를 켠 단위 테스트
make test-db      # DB 저장소 + 실제 HTTP/gRPC 전체 통합 테스트 (db, migrate 이후)
make build        # bin/에 서버 3개와 migrate 바이너리 생성
make fmt
make tidy
```

DB 테스트는 환경 변수가 없으면 skip한다. `make test-db`는 로컬 학습 DB용 DSN을 설정하며 테스트가 만든 데이터만 정리한다. 전체 통합 테스트는 임시 포트를 사용하고 생성한 서버를 모두 종료한다.

계약을 바꿀 때만 생성 도구가 필요하다.

```bash
make tools
make generate
```

생성된 `gen/`은 직접 수정하지 않는다. 대시보드는 Gateway 전용 HTTP 조합 API여서 `api/http.yaml`의 RPC 매핑과 별도로 등록한다.

## 학습 자료와 범위

[01 ORM/MySQL](docs/01-database.md) · [02 gRPC](docs/02-grpc.md) · [03 아키텍처](docs/03-architecture.md) · [04 테스트](docs/04-testing.md) · [05 Spring 비교](docs/05-spring-guide.md) · [06 읽을거리](docs/06-reading.md)

기초 문법과 worker pool은 별도 프로젝트인 [go-basic](../go-basic/README.md)에서 다룬다. 검증 방법은 [테스트 안내](docs/04-testing.md)를 본다. 사용자 인증·목록 소유권, TLS, 분산 추적 수집기, 배포 환경의 비밀 관리와 의존성 readiness는 이후 실습 범위다. `/healthz`는 프로세스 응답 여부만 확인한다.

변경 사항은 [커밋 규칙](../CONTRIBUTING.md)에 따라 기능·설정 단위로 나누어 기록한다.
