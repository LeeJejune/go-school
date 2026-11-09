# 06. 막혔을 때 어떤 문서를 읽을까요?

처음부터 공식 문서를 모두 정독할 필요는 없습니다. **저장소의 작은 코드를 읽다가 이해가 막힌 부분만 찾아보는 순서**로 사용하세요.

## 1. 지금 막힌 질문에 맞춰 골라요

| 막힌 질문 | 먼저 볼 자료 | 확인할 부분 |
| --- | --- | --- |
| :=, *, &, interface가 낯설어요 | [go-basic 학습 순서](../../go-basic/README.md) | 1~4장의 실행 예제 |
| 브라우저에서 문법을 바꿔 보고 싶어요 | [A Tour of Go](https://go.dev/tour/) | 패키지, 함수, 포인터, 메서드 |
| gRPC가 다른 서버를 어떻게 호출하나요? | [gRPC 소개](https://grpc.io/docs/what-is-grpc/introduction/) | 서비스 정의와 생성된 클라이언트 |
| HTTP JSON은 어떻게 연결되나요? | [grpc-gateway 공식 저장소](https://github.com/grpc-ecosystem/grpc-gateway) | 생성된 HTTP 처리 코드 |
| 서버 옵션과 공통 처리는 어디서 보나요? | [Kratos](https://go-kratos.dev/docs/) | HTTP/gRPC transport와 middleware |
| Fx가 생성자를 언제 호출하나요? | [Fx 시작 안내](https://uber-go.github.io/fx/get-started/) | Provide, Invoke, Lifecycle |
| GORM으로 조회·저장하는 법은요? | [GORM](https://gorm.io/docs/) | Query, Update, Transaction, Preload |
| 스키마 변경 이력을 어떻게 관리하나요? | [Goose 공식 저장소](https://github.com/pressly/goose) | SQL migration과 Provider |
| Context로 무엇이 취소되나요? | [Context 표준 문서](https://pkg.go.dev/context) | 취소 함수와 Done |
| 여러 작업의 오류를 함께 다루고 싶어요 | [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) | WithContext, Go, Wait |

링크에서 새 단어가 더 많이 나오면 이 저장소의 해당 장으로 돌아와 작은 예시를 먼저 실행해도 됩니다.

## 2. 공식 예제를 가져올 때 버전을 확인하세요

이 프로젝트가 사용하는 버전은 [go.mod](../go.mod)에 있습니다. 온라인 문서는 다른 버전의 API를 먼저 보여 줄 수 있으므로 메서드 이름과 import 경로를 함께 확인하세요.

예를 들어 이 Repository는 GORM의 db.WithContext(ctx).Create(...).Error 형태를 씁니다. 공식 문서에는 제네릭 API 예제도 있으니, 처음에는 같은 형식의 설명을 찾아 비교하는 편이 쉽습니다.

gen은 생성된 결과입니다. 이해가 안 되는 생성 코드부터 고치기보다 .proto 원본과 직접 작성한 service, UseCase, Repository를 먼저 읽으세요.

## 3. 용어를 짧게 다시 정리해요

| 용어 | 이 저장소에서의 뜻 |
| --- | --- |
| 모듈 | go.mod로 이름과 의존성을 관리하는 프로젝트 단위 |
| 패키지 | 함께 묶여 import되는 Go 코드 단위, 보통 같은 디렉터리 |
| DTO | 요청·응답 등 경계에서 데이터를 전달할 타입 |
| Entity | DB 테이블과 매핑할 구조체 |
| UseCase | 목록 생성처럼 업무 흐름을 수행하는 코드 |
| DI | 필요한 객체를 직접 찾아 만들기보다 생성자로 전달받는 방식 |
| migration | 번호를 가진 스키마 변경 이력 |
| 멱등성 | 같은 요청을 반복해도 중복 생성 같은 효과가 더 생기지 않게 하는 성질 |
| deadline | 요청이나 작업의 마감시간 |
| goroutine | Go 런타임이 관리하는 실행 흐름 |

용어를 먼저 외우기보다 실제 파일을 가리키면서 설명해 보세요. 예를 들어 '이 프로젝트에서 Controller 역할은 어디인가요?'라는 질문에 service/reading.go를 열 수 있으면 좋은 출발입니다.

## 4. 실무 과제는 하나씩 확장하세요

기초 흐름을 이해했다면 아래 순서로 한 가지씩 진행할 수 있습니다.

1. 입력 검증을 하나 추가하고 업무 단위 테스트를 작성합니다.
2. 읽음 상태 필터를 추가하고 생성되는 쿼리를 확인합니다.
3. 요청 ID를 로그에 넣어 Gateway부터 Reading까지 추적합니다.
4. DB와 다른 서버의 현재 상태를 확인하는 readiness를 추가합니다.
5. 사용자 인증과 목록 소유권, 배포 환경 설정을 설계합니다.

처음부터 프레임워크를 더 추가하는 것보다, 바꾼 기능이 어느 파일들을 지나고 어떤 테스트로 검증되는지 설명할 수 있는지 확인하세요.

[이전: 테스트](05-testing.md) · [학습 순서](../README.md)
