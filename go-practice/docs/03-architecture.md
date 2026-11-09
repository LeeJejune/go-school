# 03. 패키지 경계와 DI

## 현재 구조

```text
cmd/*                  Fx 애플리케이션 실행
internal/app           DI 모듈과 의존성 그래프
internal/server        Kratos HTTP/gRPC, 미들웨어, Fx 생명주기 연결
internal/service       RPC 요청·응답 변환과 오류 매핑
internal/reading       업무 모델, UseCase, Repository/Catalog 인터페이스
internal/catalog       자료 카탈로그 업무 기능
internal/data          GORM Entity와 Repository 구현
internal/catalogclient Reading이 호출하는 Catalog RPC 어댑터
internal/gateway       두 서버의 응답을 조합하는 BFF 기능
internal/config        환경 설정
migrations             Goose 스키마 변경
gen                    생성된 DTO와 RPC/HTTP 연결 코드
```

이 패키지 이름은 하나의 팀 관례다. Spring의 annotation 이름처럼 Go 도구가 service/data라는 이름을 인식해 자동 등록하지 않는다.

## 의존성 방향

```text
service → reading ← data
             ↑
       catalogclient

app → 이들을 Fx로 연결
server → service를 RPC 서버에 등록
```

reading은 GORM, gRPC, Fx를 import하지 않는다. 외부 호출과 저장에 필요한 작은 인터페이스만 정의한다. 새로운 저장소를 붙일 때 업무 코드보다 app의 provider와 어댑터를 바꾸는 구조다.

## Fx는 어디에서 쓰는가

`fx.Provide`는 생성 함수를 등록하고 `fx.As`는 구현체를 업무 인터페이스로 연결한다. `fx.Invoke`는 서버를 실제로 생성하게 한다. Fx가 생성자 인자를 보고 의존성 그래프를 구성한다.

```go
fx.Annotate(data.NewReadingRepository, fx.As(new(reading.Repository)))
```

Spring의 Repository 구현체를 생성자 주입으로 연결하는 역할과 대응한다. 업무 메서드 안에서 컨테이너를 조회하지 않는다. 단위 테스트에서는 Fx 없이 생성자에 대역을 전달한다.

## 생명주기

Fx가 신호와 시작·종료 순서를 소유하고 Kratos가 HTTP/gRPC transport를 소유한다. 서버 시작 훅은 포트 바인딩을 먼저 확인하고, 장시간 실행되는 Start만 goroutine으로 실행한다. Fx의 startup context는 시작 후 끝나므로 서버 전체 수명으로 재사용하지 않는다.

의존성 생성 순서로 DB/클라이언트가 먼저 준비되고 서버가 시작된다. 종료 훅은 역순으로 서버를 멈춘 뒤 연결을 닫는다. Kratos의 Stop은 제한 시간 안에서 graceful shutdown을 시도한다.

## 분리의 비용

API DTO → 업무 모델 → DB Entity 변환은 경계별 변경을 격리하지만 코드량도 늘린다. 현재는 Spring 개발자가 각 역할을 찾고 실제 도구를 이해하기 위한 명시적인 참조 구조다. 작은 읽기 전용 Catalog에는 별도 Repository를 만들지 않았다.

Gateway는 기존 API의 프로토콜 변환과 대시보드 응답 조합만 담당한다. DB 저장 규칙은 Reading에 있다. 더 큰 제품에서는 서비스 소유권, 인증, 관측 가능성과 배포 운영도 함께 설계해야 한다.

실제 요청 추적과 Spring 비교는 [실무 코드 안내](05-spring-guide.md)를 본다.
