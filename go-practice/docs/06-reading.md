# 06. 공식 문서와 읽을거리

각 단계의 코드를 먼저 훑고 해당 문서를 읽는다. 아키텍처 글은 언어와 DB를 익힌 뒤 읽는 편이 판단하기 쉽다.

| 순서 | 자료 | 볼 부분 |
| --- | --- | --- |
| 1 | [A Tour of Go](https://go.dev/tour/) | slice, map, method, interface를 직접 수정 |
| 2 | [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) | naming, error, interface, context의 관례 |
| 3 | [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines) | 받는 쪽이 종료했을 때 보내는 goroutine의 수명 |
| 4 | [Context](https://go.dev/blog/context) / [Context and structs](https://go.dev/blog/context-and-structs) | 요청 범위의 취소와 저장하면 생기는 문제 |
| 5 | [Accessing relational databases](https://go.dev/doc/database/) | database/sql, Rows, transaction, pool |
| 6 | [MySQL 드라이버](https://pkg.go.dev/github.com/go-sql-driver/mysql) | DSN, parseTime, timeout, RowsAffected |
| 7 | [gRPC Go basics](https://grpc.io/docs/languages/go/basics/) | 생성 코드와 실제 RPC 호출 |
| 8 | [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) | HTTP 매핑과 코드 생성 |
| 9 | [Go 공식 프로젝트 구조](https://go.dev/doc/modules/layout) | internal의 import 제한, 여러 cmd |
| 10 | [Three Dots Labs: Clean Architecture in Go](https://threedots.tech/post/introducing-clean-architecture/) | 실제 코드에서 의존성을 바꾸는 과정 |
| 11 | [당근: 계층형과 육각형 아키텍처 적용 사례](https://medium.com/daangn/%EC%95%84%ED%82%A4%ED%85%8D%EC%B2%98%EC%97%90-%EB%8C%80%ED%95%9C-%EA%B3%A0%EB%AF%BC%EC%9D%80-%EC%B2%98%EC%9D%8C%EC%9D%B4%EB%9D%BC-b75dffd73eb0) | 단순한 로직과 외부 의존성이 많아진 로직의 차이 |
| 12 | [당근: 예측 가능한 대규모 서비스 개발하기](https://medium.com/daangn/%EC%98%88%EC%B8%A1-%EA%B0%80%EB%8A%A5%ED%95%9C-%EB%8C%80%EA%B7%9C%EB%AA%A8-%EC%84%9C%EB%B9%84%EC%8A%A4-%EA%B0%9C%EB%B0%9C%ED%95%98%EA%B8%B0-a33e2f3cef88) | Go 서비스에서 초기 상태와 실패 경로, 측정의 중요성 |

Go 공식 문서는 API와 도구의 기준이다. 회사 글은 당시 팀의 사례다. 글의 구조를 그대로 복사하기보다 어떤 제약 때문에 해당 선택을 했는지 본다. `golang-standards/project-layout`은 이름과 달리 Go 팀의 공식 표준이 아니다.

## IDE

**GoLand**: JetBrains 사용 경험을 살리기 좋다. 탐색, 리팩터링, 테스트와 디버깅을 한 환경에서 사용한다. 상용 라이선스 또는 자격이 확인된 학생용 라이선스가 필요하다. [등록 안내](https://www.jetbrains.com/help/go/register.html), [학생 지원](https://www.jetbrains.com/academy/student-pack/)

**VS Code**: [공식 Go 확장 설정](https://code.visualstudio.com/docs/languages/go)에 따라 `golang.Go`를 설치한다. gopls가 탐색과 분석을, Delve가 디버깅을 담당한다. `.vscode/launch.json`에 서버별 시작점을 두었다.

IDE를 바꾸는 것보다 `go test`, `go vet`, race detector와 debugger를 실제로 사용하는 것이 먼저다. 생성된 파일과 작성한 파일을 구분해 읽는다.

## 이 프로젝트에서 사용하는 도구

- [Kratos v3](https://go-kratos.dev/docs/): HTTP/gRPC transport와 공통 미들웨어
- [Fx](https://uber-go.github.io/fx/): 생성자 기반 DI, 모듈, lifecycle
- [GORM](https://gorm.io/docs/): Entity, Preload, Transaction, false 업데이트
- [Goose](https://github.com/pressly/goose): 스키마 버전과 마이그레이션
- [caarlos0/env](https://github.com/caarlos0/env): 타입이 있는 환경 변수 로딩
- [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup): 동시 작업의 오류·취소·완료 대기

이 도구가 실제로 연결된 위치는 [실무 코드 안내](05-spring-guide.md)를 본다.
