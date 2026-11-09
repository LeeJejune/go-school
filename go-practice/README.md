# go-practice — Spring 개발자의 첫 Go 백엔드

Java/Spring으로 API와 DB 기능을 만들어 봤지만, Go에서는 어느 파일을 봐야 할지 모르는 사람을 위한 실습입니다. 목표는 **요청 하나가 Controller 역할의 코드에서 업무 로직과 Repository를 거치는 과정을 읽는 것**입니다.

Go 문법이 아직 낯설면 [go-basic의 1~4장](../go-basic/README.md)을 먼저 읽어 보세요. worker pool 내부 구현까지 모두 익힌 뒤 시작할 필요는 없습니다.

## 1. 먼저 JSON 응답을 확인하세요

Go 1.26 이상, Docker Desktop/Compose와 curl을 준비합니다. Docker Desktop을 켜고 편집기에서 go-practice 폴더를 여세요. 아래 명령은 모두 **go-practice 폴더**에서 실행합니다.

저장소 루트인 go-school에 있다면 다음처럼 시작합니다.

```bash
cd go-practice
make up
```

이미지를 빌드하고 MySQL → 마이그레이션 → 서버들을 시작합니다. 명령이 끝난 뒤에도 컨테이너는 실행 중입니다.

```bash
curl -sS http://localhost:8080/v1/resources
```

자료 4개의 ID·제목·URL이 들어 있는 JSON이 보이면 성공입니다. 브라우저에서도 이 주소를 열 수 있습니다. 이 실습의 결과는 웹 화면 대신 API 응답으로 확인합니다.

로그 보기와 종료 명령입니다.

```bash
docker compose logs -f catalog reading gateway
# 로그 보기만 끝낼 때 Ctrl+C

# 서버와 DB 컨테이너를 종료할 때
make down
```

make down은 DB 볼륨을 유지합니다. Compose 프로젝트 이름 go-school도 유지해 이전 학습 DB를 계속 사용할 수 있습니다.

## 2. 익숙한 역할부터 찾아보세요

| Spring에서 찾던 것 | 지금 열 파일 |
| --- | --- |
| Controller | [internal/service/reading.go](internal/service/reading.go) |
| 업무 Service | [internal/reading/usecase.go](internal/reading/usecase.go) |
| Repository 구현 | [internal/data/reading_repository.go](internal/data/reading_repository.go) |
| Entity | [internal/data/model.go](internal/data/model.go) |
| 요청·응답 DTO 원본 | [school.proto](api/proto/school/v1/school.proto) |
| DI 설정 | [internal/app/app.go](internal/app/app.go) |

폴더 이름 service만 보고 업무 로직부터 찾으면 헷갈릴 수 있습니다. 이 예제에서는 service가 Controller 역할이고, 실제 업무 흐름은 UseCase에 있습니다.

## 3. 문서는 이 순서로 읽으세요

| 순서 | 문서 | 답할 질문 |
| --- | --- | --- |
| 1 | [Spring 역할과 Go 파일 연결](docs/01-spring-to-go.md) | Controller, DTO, Service, Entity는 어디 있나요? |
| 2 | [HTTP와 gRPC 요청 흐름](docs/02-grpc.md) | Gateway, Catalog, Reading은 어떻게 호출하나요? |
| 3 | [GORM과 MySQL](docs/03-database.md) | 저장과 트랜잭션은 어디에서 하나요? |
| 4 | [DI·설정·시작과 종료](docs/04-architecture.md) | 짧은 main.go 뒤에서 누가 객체를 만드나요? |
| 5 | [테스트](docs/05-testing.md) | JUnit·Mockito 경험을 어떻게 연결하나요? |
| 6 | [필요할 때 찾는 공식 자료](docs/06-reading.md) | 막힌 부분은 무엇을 더 읽나요? |

각 장은 먼저 볼 파일, 짧은 코드 해설, 확인 문제를 포함합니다. 지금 필요한 기능부터 읽고 다시 실행해도 됩니다.

## 4. 목록을 만들고 읽음으로 바꿔 보세요

전체 서버가 실행 중인 상태에서 다른 터미널을 사용합니다.

```bash
curl -sS -X POST http://localhost:8080/v1/lists \
  -H 'Content-Type: application/json' \
  -d '{"requestKey":"first-java-go-001","title":"Go 시작하기","resourceIds":["go-tour","go-context"]}'
```

응답의 readingList.id를 복사하세요. 아래는 주요 필드만 남긴 응답 모양이며 실제 응답에는 항목과 생성 시간도 있습니다.

```json
{
  "readingList": {
    "id": "실제로 생성된 ID",
    "title": "Go 시작하기"
  }
}
```

아래 URL의 LIST_ID를 실제 ID로 교체합니다.

```bash
curl -sS http://localhost:8080/v1/lists/LIST_ID

curl -sS -X PATCH http://localhost:8080/v1/lists/LIST_ID/items/go-tour \
  -H 'Content-Type: application/json' \
  -d '{"isRead":true}'

curl -sS http://localhost:8080/v1/lists/LIST_ID
```

PATCH의 성공 응답이 {}여도 정상입니다. 다시 조회한 목록에서 go-tour 항목의 isRead가 true인지 확인하세요. false로 보내면 미읽음으로 바꿉니다.

같은 requestKey와 같은 내용으로 POST를 반복하면 같은 ID를 반환합니다. 같은 키에 다른 제목이나 자료 순서를 보내면 409입니다. 새 목록을 만들 때는 first-java-go-002처럼 새 키를 사용하세요.

두 서버를 동시에 조회해 응답을 합치는 API도 있습니다.

```bash
curl -sS http://localhost:8080/v1/dashboard/LIST_ID
```

catalog와 readingList가 함께 나옵니다. 이 API의 실제 goroutine 활용은 [2장](docs/02-grpc.md)에서 읽습니다.

## 5. Go 코드를 직접 디버깅하고 싶다면

앞에서 전체 Docker 실행을 사용했다면 먼저 make down으로 같은 포트의 서버를 종료합니다. 이후 MySQL만 Docker에 두고 Go 서버는 직접 실행합니다.

```bash
make deps
make db
make migrate
make migration-status
```

터미널 세 개를 열고, **각 터미널을 go-practice 폴더로 이동한 뒤** 각각 실행합니다.

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

로컬 서버 명령은 종료하지 않고 요청을 기다리는 것이 정상입니다. API 요청은 별도 터미널에서 보냅니다. 각 서버는 Ctrl+C로 종료하고 DB는 make down으로 종료합니다.

VS Code에서는 이 폴더의 .vscode/launch.json에 있는 서버별 구성을 선택할 수 있습니다. DB와 마이그레이션을 먼저 준비한 뒤 CreateList에 중단점을 걸어 보세요.

## 6. 지금 사용하는 도구

| 도구 | 담당 영역 |
| --- | --- |
| Kratos v3 | HTTP/gRPC 서버·클라이언트와 공통 처리 |
| Fx | DI와 시작·종료 관리 |
| GORM + MySQL | Entity 매핑과 DB 접근 |
| Goose | 스키마 변경 이력 |
| caarlos0/env | 환경 변수를 설정 구조체로 읽기 |
| Protobuf + grpc-gateway | 통신 DTO·RPC 계약과 HTTP 변환 |
| errgroup | 독립적인 두 조회의 동시 실행·오류 처리 |

버전은 [go.mod](go.mod)에 고정합니다. 이 조합은 하나의 참조 구성입니다. 사용자 인증·목록 소유권·TLS 등은 이후 실습 범위입니다. /healthz는 프로세스 응답 여부를 확인하는 경로이며 모든 의존성의 정상 동작까지 확인하지는 않습니다.

기본 설정으로 바로 시작할 수 있습니다. 설정을 바꾸려면 [4장](docs/04-architecture.md)을 읽으세요. .env는 자동으로 읽지 않습니다.

## 7. 자주 쓰는 검사 명령

```bash
make test
make check

# 실제 DB 테스트는 DB와 스키마를 준비한 뒤 실행
make db
make migrate
make test-db
```

실행만 할 때는 생성 도구가 필요 없습니다. .proto나 HTTP 매핑을 수정할 때만 make tools와 make generate를 사용합니다. gen은 직접 수정하지 않습니다.

명령이 없다고 나오면 현재 폴더부터 확인하세요. 저장소 루트에서는 make -C go-practice up처럼 -C로 대상을 지정할 수 있습니다.

[저장소 전체 안내](../README.md) · [기초 Go 실습](../go-basic/README.md) · [커밋 규칙](../CONTRIBUTING.md)
