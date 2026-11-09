# 05. JUnit과 Mockito 경험으로 Go 테스트 읽기

이 장의 목표는 **DB 없이 확인할 테스트와 실제 DB가 필요한 테스트를 구분하는 것**입니다. 모든 명령은 go-practice 폴더에서 실행합니다.

## 1. 우선 DB 없이 실행하세요

```bash
make test
```

테스트가 통과하면 ok가 표시됩니다. 코드가 없는 테스트 대상 패키지는 [no test files]로 표시될 수 있습니다.

검사를 조금 더 포함하려면 다음을 실행합니다.

```bash
make check
```

이 명령은 go vet라는 정적 검사와 go test -race ./...를 실행합니다. -race는 실행된 코드에서 잘못된 동시 메모리 접근을 찾는 검사입니다. 모든 업무상 경쟁 조건을 해결해 주는 기능은 아닙니다.

## 2. @Test 대신 이름과 파일 규칙을 써요

[internal/reading/usecase_test.go](../internal/reading/usecase_test.go)를 여세요.

```go
func TestCreateValidationAndNormalization(t *testing.T)
```

실제 함수의 선언 부분입니다.

| Java에서 익숙한 것 | Go에서 읽을 것 |
| --- | --- |
| 테스트 클래스·@Test | _test.go 파일과 Test로 시작하는 함수 |
| assertEquals / assertTrue | 값 비교 후 t.Fatalf나 t.Fatal로 실패 보고 |
| 파라미터 테스트 | 여러 입력을 반복하거나 t.Run으로 하위 테스트 실행 |
| Mock / Stub | 필요한 인터페이스를 구현한 작은 대역 |

특정 테스트만 보고 싶으면 Go 명령으로 실행할 수도 있습니다.

```bash
go test ./internal/reading -run TestCreateValidationAndNormalization -v
```

-v는 테스트 이름과 상세 결과를 보여 줍니다. 이 테스트는 MySQL이나 실행 중인 서버가 필요 없습니다.

## 3. Mockito 없이 대역을 넣을 수 있어요

업무 코드가 의존하는 Repository는 Create, Get, SetRead 메서드를 요구합니다. 테스트 파일의 stubRepository가 이 메서드들을 구현합니다. DB에 저장하는 대신 전달된 CreateParams를 기억합니다.

테스트에서 다음처럼 직접 생성합니다.

```go
repo := &stubRepository{}
u := New(repo, stubCatalog{}, 300*time.Millisecond)
```

실제 테스트의 일부입니다. Fx 컨테이너를 띄우지 않고 생성자에 대역을 넣었습니다. 업무 로직은 구체적인 GORM 객체를 요구하지 않기 때문에 이런 테스트가 가능합니다.

TestCatalogFailureDoesNotWrite를 읽으면 Catalog 조회가 실패했을 때 저장 호출을 하지 않았는지 확인합니다. '정상 결과가 나오는가'뿐 아니라 **실패 뒤에 하면 안 되는 일을 하지 않았는가**도 검사합니다.

## 4. 어느 테스트 파일부터 읽나요?

| 읽을 파일 | 확인할 질문 |
| --- | --- |
| [reading/usecase_test.go](../internal/reading/usecase_test.go) | 빈 값·중복 입력을 거절하고 저장 전에 검증하는가? |
| [service/reading_test.go](../internal/service/reading_test.go) | 업무 오류가 올바른 gRPC 코드로 바뀌는가? |
| [app/app_test.go](../internal/app/app_test.go) | Fx에 등록한 의존성들이 빠짐없이 연결되는가? |
| [config/config_test.go](../internal/config/config_test.go) | 잘못된 제한 시간이나 풀 크기를 거절하는가? |
| [gateway/dashboard_test.go](../internal/gateway/dashboard_test.go) | 두 조회를 동시에 시작하고, 실패하면 취소·완료를 기다리는가? |

Fx 그래프 검사는 실제 DB 연결이나 서버 시작까지 검증하는 테스트가 아닙니다. 대상별로 확인하는 경계가 다릅니다.

## 5. 실제 MySQL이 필요한 테스트도 실행해요

```bash
make db
make migrate
make test-db
```

이 순서는 MySQL 시작 → 스키마 준비 → 통합 테스트입니다. Docker Desktop이 실행 중이어야 합니다.

| 테스트 | 실제로 사용하는 것 |
| --- | --- |
| [data/reading_repository_integration_test.go](../internal/data/reading_repository_integration_test.go) | GORM과 로컬 MySQL |
| [app/integration_test.go](../internal/app/integration_test.go) | HTTP Gateway, 두 gRPC 서버, GORM, MySQL |

저장소 테스트는 같은 키의 동시 생성, 내용 충돌, 항목 저장 실패 시 rollback, true/false 반복 변경을 확인합니다. 전체 API 테스트는 생성·조회·변경과 HTTP 오류 상태, Catalog 장애 상황까지 확인합니다.

전체 API 테스트는 테스트 프로세스 안에서 세 서버를 실제 TCP 임시 포트로 실행합니다. 이미 실행 중인 8080/9001/9002 서버를 빌려 쓰지 않습니다. 끝나면 서버를 종료하고 테스트가 만든 키의 DB 행만 정리합니다.

MYSQL_TEST_DSN이 없으면 이 두 테스트는 skip합니다. make test-db는 로컬 학습 DB의 DSN을 설정합니다. 따라서 make test가 통과했다는 사실만으로 실제 MySQL 테스트까지 실행됐다고 판단하면 안 됩니다.

## 6. 동시성 테스트는 시간만 비교하지 않아요

대시보드 테스트에서는 두 조회가 서로 '시작했다'는 신호를 기다리게 합니다. 구현이 실수로 순차 호출로 바뀌면 한쪽이 다른 쪽의 시작을 기다리다 실패합니다.

이 방식은 '몇 ms 이내면 통과'처럼 컴퓨터 속도에 민감한 검사보다, **정말 두 작업을 겹쳐 실행했는가**를 직접 확인합니다.

실패 테스트는 한 조회가 오류를 내면 다른 조회가 취소 신호를 받는지, 그 작업이 끝나기 전에 Get이 반환하지 않는지도 확인합니다.

## 직접 해볼 과제

처음에는 아래에서 하나만 골라 보세요. 구현을 바꾸는 실험은 끝나면 복구합니다.

1. usecase_test.go의 입력 제목을 바꾸고 정규화 결과를 예상해 보세요.
2. Repository 대역이 받은 CreateParams를 찾아 실제로 어떤 값이 저장될지 읽어 보세요.
3. dashboard_test.go에서 시작 신호와 취소 신호를 나타내는 channel을 찾아보세요.

DB가 준비되지 않았는데 통합 테스트를 실행하면 연결 실패가 납니다. 먼저 make db와 make migrate가 성공했는지 확인하세요. 주소가 이미 사용 중이라는 메시지는 같은 포트의 다른 프로세스를 확인하는 단서입니다.

[이전: DI와 설정](04-architecture.md) · [다음: 필요한 공식 문서 찾기](06-reading.md)
