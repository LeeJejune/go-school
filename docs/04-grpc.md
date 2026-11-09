# 04. gRPC와 grpc-gateway

## 왜 두 서버인가

Catalog는 자료 정보, Reading은 목록과 읽음 상태를 맡는다. 둘로 분리하면 네트워크 호출의 지연, 취소와 실패가 실제 경계에 생긴다. 같은 함수 호출처럼 성공을 가정할 수 없다는 것을 배우기 위한 구성이다. 서버 프로세스를 두 개로 나누는 것과 실제 머신 두 대를 준비하는 것은 다르다. 이 저장소는 별도 프로세스와 컨테이너 구성을 제공하며 원격 인프라를 생성하지 않는다.

## Protobuf

`.proto`는 서비스 메서드와 메시지를 정의하는 계약이다. 필드의 숫자는 wire format에서 식별자로 사용된다. 이름을 바꾸는 일과 번호를 바꾸는 일의 영향이 다르다. 한 번 쓴 필드를 지웠다면 그 번호를 다른 의미로 재사용하지 않고 reserved로 남긴다.

proto3 scalar는 기본값을 가진다. bool의 미설정과 false를 구분해야 하면 optional 또는 별도의 업데이트 계약이 필요하다. 여기서 SetRead는 지정한 상태로 맞추는 명령이며 `isRead` 생략은 false와 같다. 일반적인 부분 수정 API의 field mask 예제는 아니다.

생성 파일은 메시지 타입, 서버 interface, 클라이언트 stub과 Gateway 변환 코드를 제공한다. 업무 검증과 SQL을 대신 작성하지 않는다.

## 호출 흐름

목록 생성 요청은 다음 순서다.

1. Gateway가 HTTP JSON을 Protobuf 요청으로 변환
2. 생성된 gRPC 클라이언트가 Reading 호출
3. Reading이 입력을 검증하고 Catalog를 한 번에 조회
4. 자료의 제목과 URL을 받아 MySQL 트랜잭션으로 저장
5. 결과가 gRPC 응답 → JSON 응답으로 반환

자료마다 RPC를 한 번씩 보내는 대신 BatchGetResources로 최대 20개를 한 번에 조회한다. 읽음 상태 변경과 저장된 목록 조회는 Catalog를 다시 호출하지 않는다.

## Gateway가 하는 일

grpc-gateway는 HTTP/JSON 클라이언트가 기존 gRPC API를 사용할 수 있게 변환한다. 여기서는 `.proto` annotation 대신 `api/http.yaml`에 HTTP 매핑을 분리했다. `Register...HandlerFromEndpoint`는 실제 RPC 호출을 연결한다. `Register...HandlerServer`로 서버 함수를 직접 호출하는 형태와 차이가 있다.

JSON 필드는 기본적으로 lowerCamelCase다. 요청 예시는 `requestKey`, Go 생성 필드는 `RequestKey`, proto 선언은 `request_key`다. 생성 코드의 규칙을 그대로 따른다.

## deadline과 오류

`grpc.NewClient`가 성공했다는 것은 원격 서버 연결이 완료됐다는 뜻이 아니다. 실제 RPC에서 연결 실패가 드러날 수 있다. 전체 요청에는 2초 예산을 주고, Catalog 조회는 그 안에서 최대 300ms를 사용한다. 상위 요청 시간이 먼저 끝나면 그 deadline이 우선한다.

| gRPC 코드 | 의미 | 일반적인 Gateway HTTP 매핑 |
| --- | --- | --- |
| InvalidArgument | 입력 잘못됨 | 400 |
| NotFound | 목록/항목 없음 | 404 |
| AlreadyExists | 같은 요청 키에 다른 내용 | 409 |
| Unavailable | 카탈로그 접근 실패 | 503 |
| DeadlineExceeded | 요청 시간 초과 | 504 |
| Internal | 예상하지 못한 저장 오류 | 500 |

원본 SQL 오류와 계정 정보를 외부에 그대로 내보내지 않는다. 서버 로그에는 원인을 남긴다. 예제는 Catalog의 짧은 조회 제한 초과를 dependency 실패로 다루어 Unavailable을 반환하고, 전체 요청 취소는 상위 context 오류를 반환한다. 같은 키로 안전하게 재시도할 수 있다는 사실과 모든 오류를 무한 재시도해도 된다는 주장은 다르다.

## 종료와 health

SIGINT/SIGTERM을 받으면 새 일을 받지 않고 진행 중 RPC의 완료를 기다린다. 3초 안에 끝나지 않으면 강제 Stop으로 넘어간다. HTTP 서버도 Shutdown 예산을 둔다. goroutine을 만든 쪽에서 끝나는 조건까지 관리한다.

Gateway의 `/healthz`는 프로세스 응답 여부다. DB와 Catalog의 현재 정상 동작을 보장하는 readiness는 아니다. 각 gRPC 서버에는 기본 health와 reflection을 등록했지만 운영 환경에서는 공개 범위와 readiness를 별도로 정해야 한다.

## 환경 변수

| 변수 | 기본값 | 프로세스 |
| --- | --- | --- |
| HTTP_ADDR | 127.0.0.1:8080 | Gateway |
| CATALOG_ADDR | 127.0.0.1:9001 | Catalog |
| READING_ADDR | 127.0.0.1:9002 | Reading |
| CATALOG_TARGET | 127.0.0.1:9001 | Reading, Gateway |
| READING_TARGET | 127.0.0.1:9002 | Gateway |
| MYSQL_DSN | README의 로컬 DB | Reading |

DSN 예시:

```bash
export MYSQL_DSN='school:school@tcp(127.0.0.1:3307)/school?parseTime=true&loc=UTC&timeout=3s&readTimeout=3s&writeTimeout=3s'
```

로컬 gRPC는 TLS 없이 연결한다. 사용자 인증과 권한도 없다. 운영에서는 TLS 또는 mTLS, 인증 주체와 소유권 검사, 요청 제한, tracing, 재시도 정책을 함께 설계한다.

읽기: [gRPC Go basics](https://grpc.io/docs/languages/go/basics/), [Deadlines](https://grpc.io/docs/guides/deadlines/), [Gateway 외부 매핑](https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/grpc_api_configuration/).
