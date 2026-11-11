# 04. Spring Boot가 하던 조립을 Fx와 Gin·grpc-go에서 찾기

이 장의 목표는 **main.go가 무엇을 시작하는지, Repository 구현체가 어떻게 주입되는지 찾는 것**입니다. 앞 장까지 요청 처리와 저장 코드를 봤으므로 이제 그 객체들이 어디에서 만들어지는지 확인합니다.

## 1. 도구 세 개를 구분하세요

| 도구 | 맡긴 역할 |
| --- | --- |
| Fx | 객체 생성자 연결, 시작·종료 훅, 프로세스 신호 처리 |
| Gin + grpc-go | HTTP 라우터와 RPC 서버·클라이언트 |
| go-grpc-middleware | RPC 로그·panic 복구·클라이언트 제한 시간 |
| GORM | DB 매핑과 SQL 실행 |

Spring Boot가 여러 기능을 한 체계로 제공하던 것과 비교하면, 이 프로젝트는 역할별 도구를 선택해 연결한 구성입니다. 이 조합이 모든 Go 팀의 정답은 아닙니다. 지금은 실제 조립 위치를 읽는 것이 목표입니다.

## 2. 시작점에서 DI 설정으로 이동하세요

[cmd/reading/main.go](../cmd/reading/main.go)를 열면 app.Reading을 Fx로 실행합니다. [internal/app/app.go](../internal/app/app.go)의 Reading 함수를 찾으세요.

```go
fx.Provide(
    mysqlrepo.NewDB,
    fx.Annotate(mysqlrepo.NewReadingRepository, fx.As(new(reading.Repository))),
    fx.Annotate(catalogclient.New, fx.As(new(reading.Catalog))),
    // UseCase, 요청 처리기, 서버의 생성자도 등록합니다.
)
```

실제 provider 목록에서 중요한 부분만 떼어 낸 코드입니다.

fx.Provide는 '이 함수를 사용해 필요한 객체를 만들 수 있다'고 등록합니다. 함수가 요구하는 인자도 등록된 생성자에서 찾아 연결합니다. 등록한 모든 생성자를 즉시 실행한다기보다는 필요한 객체에서부터 의존성을 따라갑니다.

fx.Invoke(func(*server.GRPCServer) {})는 서버 객체가 필요하다고 명시하여 이 생성 과정을 시작합니다. 빈 함수 본문보다 인자로 요구하는 서버 타입이 중요한 부분입니다.

## 3. 인터페이스를 어떻게 주입하나요?

업무 코드에는 다음 계약이 있습니다.

```go
type Repository interface {
    Create(context.Context, CreateParams) (List, error)
    Get(context.Context, string) (List, error)
    SetRead(context.Context, string, string, bool) error
}
```

[reading/model.go](../internal/reading/model.go)에 정의된 인터페이스입니다. Go에서는 구현체가 메서드를 제공하면 이 계약을 만족합니다. Fx에는 어떤 구현체를 쓸지 다음처럼 알려 줍니다.

```go
fx.Annotate(mysqlrepo.NewReadingRepository, fx.As(new(reading.Repository)))
```

'NewReadingRepository가 만든 객체를 reading.Repository로 제공하라'는 뜻입니다. new(reading.Repository)는 여기서 Fx에 인터페이스 타입을 알려 주기 위해 사용합니다. 이것만으로 DB Repository 구현체가 생성되는 것은 아닙니다.

Spring의 @Bean 메서드가 인터페이스 타입을 반환하고, Service 생성자가 그 인터페이스를 주입받는 상황을 떠올리면 됩니다.

## 4. UseCase는 컨테이너를 모르고 일해요

[reading.New](../internal/reading/usecase.go)는 Repository, Catalog, 조회 제한 시간을 받습니다. 이 함수 안에서 Fx를 조회하거나 GORM 연결을 만들지 않습니다.

```text
RPC 처리기 → reading.UseCase → reading.Repository 약속
                                 ↑
                    mysql.ReadingRepository가 구현
```

UseCase는 '저장할 수 있는 대상'을 요청하고, 구현체 선택은 app 모듈에서 합니다. 이 때문에 업무 테스트에서는 MySQL 대신 약속을 만족하는 작은 대역을 넣을 수 있습니다.

Repository 인터페이스가 업무 패키지에 있는 이유도 같습니다. 실제 DB의 모든 기능을 나열한 것이 아니라, 업무 코드가 필요로 하는 세 동작만 적었습니다.

## 5. Spring 설정 파일에 해당하는 곳

[internal/config/config.go](../internal/config/config.go)의 구조체 태그를 caarlos0/env 라이브러리가 읽습니다.

```go
RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" envDefault:"2s"`
```

'REQUEST_TIMEOUT 환경 변수를 읽고, 없으면 2초를 사용한다'는 뜻입니다. Spring의 @ConfigurationProperties로 설정을 타입에 바인딩하던 역할과 연결할 수 있습니다.

| 설정 | 기본값 | 바꾸는 이유 |
| --- | --- | --- |
| HTTP_ADDR | 127.0.0.1:8080 | Gateway 수신 주소 |
| CATALOG_ADDR / READING_ADDR | 127.0.0.1:9001 / 127.0.0.1:9002 | 각 RPC 서버 수신 주소 |
| CATALOG_TARGET / READING_TARGET | 127.0.0.1:9001 / 127.0.0.1:9002 | 호출할 다른 서버 주소 |
| MYSQL_DSN | 로컬 school DB 연결 문자열 | MySQL 접속 대상 |
| REQUEST_TIMEOUT | 2s | 요청 제한 시간 |
| CATALOG_TIMEOUT | 300ms | Reading의 자료 조회 제한 |
| DB_MAX_OPEN / DB_MAX_IDLE | 10 / 5 | 연결 풀 크기 |

기본값으로 실습할 때 설정 파일을 만들 필요는 없습니다. 예를 들어 Gateway 포트만 바꾸려면 go-practice 폴더에서 다음처럼 실행합니다.

```bash
HTTP_ADDR=127.0.0.1:8081 make run-gateway
```

이 경우 HTTP 요청도 8081로 보내야 합니다. 위 셸 명령은 macOS/Linux 기준입니다.

.env.example을 복사해 .env를 만들 수는 있지만 자동 로딩되지는 않습니다. 현재 셸에 읽고 하위 프로세스에 넘기려면 다음처럼 사용합니다.

```bash
cp .env.example .env
set -a
source .env
set +a
```

여러 터미널에서 로컬 서버를 실행한다면 각 셸에 필요한 환경 변수를 적용해야 합니다. Docker Compose의 컨테이너 설정은 [compose.yaml](../compose.yaml)에서 따로 확인하세요.

## 6. 시작·종료 순서는 왜 필요한가요?

서버가 시작되려면 필요한 DB와 클라이언트가 준비되어야 합니다. 종료할 때는 진행 중인 요청이 사용하는 연결을 너무 일찍 닫으면 안 됩니다.

[DB 생성 코드](../internal/repository/mysql/database.go)는 Fx의 OnStart에 연결 확인, OnStop에 연결 종료를 등록합니다. [server/lifecycle.go](../internal/server/lifecycle.go)는 HTTP/gRPC 서버를 Fx 생명주기에 연결합니다.

```text
시작: DB 연결 확인 → HTTP/gRPC 서버 시작
종료: HTTP/gRPC 서버 중지 → 클라이언트와 DB 연결 닫기
```

Fx가 Ctrl+C 같은 종료 신호를 처리하고, 종료 훅은 역순으로 실행합니다. HTTP/gRPC 서버는 등록된 종료 훅에서 정리합니다.

서버의 Start는 오래 실행되므로 goroutine에서 진행하고 종료를 기다립니다. Fx가 시작 작업에 준 Context는 시작 이후의 서버 전체 수명과 다르기 때문에 그대로 재사용하지 않는다는 주석도 찾아보세요.

## 7. Spring Boot보다 초기 조립이 더 보이는 이유

Gin은 라우팅·바인딩·미들웨어를 제공하는 HTTP 프레임워크입니다. DI 컨테이너나 ORM, migration까지 한 체계로 자동 구성하지 않습니다. 이 프로젝트에서는 Fx·GORM·Goose로 그 역할을 연결했습니다.

`cmd/gateway/main.go`는 진입점이고, DI 컨테이너 역할은 Fx, 등록 설정은 `internal/app/app.go`입니다. 업무 객체는 생성자로 의존성을 받으므로 DI 도구 없이 테스트에서도 직접 만들 수 있습니다.

이 선택은 부품을 직접 연결하는 만큼 설정 위치를 읽어야 합니다. 팀이 더 많은 규약·코드 생성을 원하면 Kratos 같은 서비스 프레임워크도 비교할 수 있습니다. Gin을 골랐다는 사실만으로 애플리케이션 전체 구조가 결정되지는 않습니다. [프로젝트 비교](07-project-comparison.md)에 선택 이유를 적었습니다.

## 8. 여러 요청이 같은 객체를 써도 되나요?

Spring의 singleton Service가 여러 요청을 받던 것처럼 UseCase와 Repository 객체도 여러 요청에서 사용됩니다. 요청별 제목이나 ID는 함수의 지역 변수와 인자로 다룹니다. 공유 객체의 필드에 이번 요청 데이터를 저장하면 동시 요청이 서로 덮어쓸 수 있습니다.

현재 UseCase 필드에는 Repository와 Catalog, 제한 시간처럼 공유할 의존성·설정만 있습니다. 각 메서드가 요청의 Context를 받는 것도 이 구분과 연결됩니다.

## 확인 문제

1. Repository 구현체를 바꾸려면 먼저 어느 파일의 등록을 보나요?
2. envDefault 태그 자체는 Go 문법인가요, 값을 읽는 동작은 누가 하나요?
3. 서버가 시작될 때 받은 모든 Context를 서버 종료 때까지 써도 되나요?

답: internal/app/app.go를 봅니다. 태그는 구조체 메타데이터이고 env 라이브러리가 읽습니다. Context의 수명과 용도가 달라 시작 작업과 요청·서버 수명을 구분해야 합니다.

[이전: GORM과 DB](03-database.md) · [다음: 테스트](05-testing.md)

더 읽기: [Fx 시작 안내](https://uber-go.github.io/fx/get-started/), [Gin](https://gin-gonic.com/en/docs/), [grpc-go](https://github.com/grpc/grpc-go).
