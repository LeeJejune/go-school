# 07. 국내외 Go 프로젝트를 비교하고 구조 고르기

확인일: **2026-11-11**. 아래는 공개 저장소의 README·의존성·실제 서버 코드를 비교한 결과입니다. 공개된 한 프로젝트를 회사 전체의 개발 표준으로 일반화하지 않습니다. 기본 브랜치의 내용은 이후 바뀔 수 있습니다.

## 1. 어떤 프로젝트들을 봤나요?

| 프로젝트 | 공개 코드에서 확인한 사실 | 이 실습에 적용한 판단 |
| --- | --- | --- |
| [당근 accesslog](https://github.com/daangn/accesslog/blob/main/README.md) | HTTP 접근 로그 라이브러리. README는 chi와 zerolog를 조합한 예를 제공 | 요청 메서드·경로·상태·지연을 공통 미들웨어로 기록. 라이브러리여서 전체 CRUD 구조의 기준으로 쓰지는 않음 |
| [네이버 lobster](https://github.com/naver/lobster) | 분산 로그 수집·검색 시스템. [의존성](https://github.com/naver/lobster/blob/main/go.mod)에 gorilla/mux와 grpc-go, [설계 문서](https://github.com/naver/lobster/blob/main/docs/design/README.md)에 query/store 등 역할 분리 | 프로세스·업무별 책임을 구분. 로그 청크 저장소와 Kubernetes 구성은 이 실습의 CRUD 규모에 맞춰 가져오지 않음 |
| [카카오 varlog](https://github.com/kakao/varlog/blob/main/README.md) | 강한 일관성을 제공하는 분산 로그 저장소. [의존성](https://github.com/kakao/varlog/blob/main/go.mod)에 grpc-go 등 사용 | RPC 계약과 저장소 책임을 구분하는 비교 대상. 복제·합의·로그 스트림 구조는 일반 웹 CRUD와 목적이 다름 |
| [Yorkie](https://github.com/yorkie-team/yorkie) | 실시간 협업 문서 서버. [RPC 서버](https://github.com/yorkie-team/yorkie/blob/main/server/rpc/server.go)는 ConnectRPC와 net/http, [Backend](https://github.com/yorkie-team/yorkie/blob/main/server/backend/backend.go)는 DB 등 자원 관리, [Database 계약](https://github.com/yorkie-team/yorkie/blob/main/server/backend/database/database.go)은 구현 분리 | 통신 계층과 업무·저장을 나누고, 서버 바인딩 실패를 시작 시점에 확인. CRDT와 거대한 DB 인터페이스는 복제하지 않음 |
| [Evrone go-clean-template](https://github.com/evrone/go-clean-template) | controller / usecase / repo 경계와 생성자 조립을 설명. **확인 시점 REST 구현은 Fiber** | Controller → 업무 흐름 → Repository의 경계를 참고. 메시지 브로커·전체 템플릿을 복사하지 않고 HTTP 구현은 Gin 선택 |
| [Fn Project](https://github.com/fnproject/fn/blob/master/api/server/server.go) | 실제 서버 코드에 Gin을 사용하는 함수 실행 플랫폼 | Gin이 템플릿 외 실제 서버에도 쓰이는 사례로 확인. 함수 실행 엔진 구조까지 CRUD에 가져오지는 않음 |

표의 마지막 열은 **이 프로젝트의 규모와 학습 목적에 따른 판단**입니다. 참고한 것은 책임 분리와 실행·종료 패턴이며, 외부 코드를 통째로 복사하지 않았습니다.

당근 공개 라이브러리의 chi 예시만으로 '당근은 Gin을 안 쓴다'고 결론 내릴 수도 없습니다. 마찬가지로 네이버·카카오의 인프라 저장소를 웹 백엔드 표준 템플릿으로 취급하면 맞지 않습니다.

## 2. Gin을 선택한 근거

[JetBrains Developer Ecosystem 2026 자료](https://blog.jetbrains.com/go/2026/10/13/python-vs-go/)의 Go 웹 도구 정기 사용 응답은 Gin 48%, net/http 46%, gorilla/mux·chi 각각 15%, Echo·Fiber 각각 13%입니다. 해당 질문 응답자 1,519명, 가중 데이터, 복수 응답입니다. 국내 기업만의 점유율이 아니고 Kratos 개별 수치도 없어 Gin과 Kratos의 정확한 국내 비율을 계산할 수는 없습니다.

이 실습에서는 사용 사례가 넓고 URL·DTO·Controller를 직접 따라가기 쉬운 **Gin**을 HTTP 계층에 선택했습니다. Gin은 HTTP 프레임워크이고 Kratos는 HTTP/gRPC와 서비스 개발 규약까지 다루므로 단순히 같은 기능의 대체품이라고만 볼 수는 없습니다. [Gin 문서](https://gin-gonic.com/en/docs/)와 [Kratos 저장소](https://github.com/go-kratos/kratos)를 비교해 보세요.

gRPC에는 grpc-go, DI에는 Fx, ORM에는 GORM, migration에는 Goose를 사용합니다. 이것이 국내 회사들의 단일 표준 조합이라는 뜻은 아닙니다. HTTP 프레임워크를 바꾸더라도 업무 로직을 유지할 수 있도록 경계를 둡니다.

## 3. Controller-Service-Repository와 Clean Architecture 중 무엇이 맞나요?

둘 중 이름 하나를 골라 모든 디렉터리를 맞추기보다, 이 실습에 필요한 의존성을 정했습니다.

```text
cmd/<앱>/main.go                   실행 진입점
internal/app                      생성자 DI와 앱 조립
internal/controller/http          Gin 라우팅·HTTP DTO·응답 처리
internal/controller/grpc          RPC 요청·업무 모델 변환
internal/catalog                  자료 업무 규칙·Repository 계약
internal/reading                  목록 업무 규칙·Repository/Catalog 계약
internal/repository/mysql         GORM 구현과 DB Entity
internal/catalogclient            Reading용 Catalog RPC 어댑터
internal/gateway                  두 RPC 응답의 조합
internal/rpc                      공통 RPC 클라이언트·interceptor
internal/server                   수신·실행·정상 종료
api/proto                         서버 간 통신 계약 원본
gen                               생성된 RPC 타입·연결 코드
migrations                        스키마 변경 이력
requests                          수동 HTTP·Postman 입력
```

Spring 역할로 읽으면 Controller-Service-Repository입니다. 업무가 구체적인 통신·DB 라이브러리를 모르고 작은 인터페이스를 요구한다는 점은 Clean Architecture에서 참고한 부분입니다.

| 선택 | 이유 |
| --- | --- |
| 업무는 catalog·reading처럼 기능별로 모음 | 새 업무가 생겼을 때 관련 규칙과 계약을 가까이 읽기 |
| Controller는 HTTP/gRPC별로 분리 | 바인딩과 오류·응답 변환을 업무에서 분리 |
| Repository 인터페이스는 업무 쪽에 둠 | 소비자가 필요한 작은 계약만 요구하고 테스트 대역 교체 |
| DB Entity는 mysql 안에 둠 | ORM 태그와 테이블 변경이 업무 모델·외부 DTO에 퍼지지 않도록 함 |
| 모든 CRUD를 감싸는 범용 BaseRepository 생략 | 현재 업무에 없는 추상화·옵션 조합을 늘리지 않음 |
| main은 짧게, DI는 app에서 명시 | 등록 위치를 찾기 쉽고 업무 테스트에서 컨테이너 없이 생성 |

Go에는 강제되는 하나의 Controller-Service-Repository 디렉터리 표준이 없습니다. 공식 [서버 프로젝트 구성 안내](https://go.dev/doc/modules/layout#server-project)도 cmd와 internal을 포함한 구성을 설명하지만 모든 업무 계층 이름을 지정하지는 않습니다.

## 4. 어느 정도까지 실무에 가까운가요?

현재 코드에서 직접 연습하는 것은 입력 검증, 오류 변환, 요청 취소·제한 시간, DB 풀, 트랜잭션·멱등성, 정상 종료, 동시 RPC 조회, 실제 DB·전체 API 테스트입니다. 자료 샘플은 서버에 내장하지 않고 HTTP 요청으로 등록합니다.

다음 단계는 사용자 인증·소유권, 페이징, 관측성, 비밀 설정·TLS, 배포·운영 정책입니다. 분산 로그 프로젝트를 참고했다는 이유만으로 큐나 Kubernetes를 먼저 추가하지 않았습니다. 실제 업무와 트래픽에 맞춰 확장하면서 각 경계의 테스트를 유지하는 것이 이 실습의 다음 과제입니다.

[이전: 공식 자료](06-reading.md) · [실행과 학습 순서](../README.md)
