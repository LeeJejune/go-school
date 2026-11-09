# 04. 테스트와 다음 단계

## 테스트 경계

| 코드 | 확인하는 동작 |
| --- | --- |
| internal/reading/usecase_test.go | 검증, 제목 정규화, fingerprint, 외부 조회 실패 시 저장 금지 |
| internal/service/reading_test.go | 업무 오류의 gRPC 매핑, 내부 오류 노출 방지 |
| internal/gateway/dashboard_test.go | 실제 동시 시작, 한쪽 실패 시 취소·완료 대기, HTTP 상태·JSON 필드 |
| internal/app/app_test.go | 세 애플리케이션의 Fx 의존성 그래프 |
| internal/config/config_test.go | 유효하지 않은 제한 시간과 DB 연결 풀 설정 거절 |
| internal/data/reading_repository_integration_test.go | 실제 MySQL에서 동시 같은 키 생성, 충돌, 롤백, true/false 반복 변경 |
| internal/app/integration_test.go | 실제 TCP HTTP → Gateway → gRPC → GORM → MySQL, Catalog 장애, 종료 |

```bash
make check
make db
make migrate
make test-db
```

`make check`는 vet와 race detector를 실행한다. DB 관련 테스트는 MYSQL_TEST_DSN이 없으면 skip하며 `make test-db`가 로컬 DSN을 설정한다. 단위 테스트 통과와 MySQL 통합 테스트 통과를 구분한다.

DB 테스트는 테스트가 만든 키의 행만 정리한다. 전체 API 테스트는 서버 세 개를 임시 포트에서 시작하고 종료한다. 고정 sleep이나 기존 8080/9001/9002 포트에 의존하지 않는다.

동시성 테스트는 작업 둘이 서로 시작 신호를 확인하게 하므로, 실행 시간을 비교하는 테스트보다 순차 구현으로 퇴행했는지를 명확히 확인한다.

## 직접 확인할 실패 경로

1. 같은 requestKey, 같은 내용으로 생성 2번: ID 동일
2. 같은 requestKey에 제목 변경: 409
3. 없는 자료 선택: 400, DB 저장 없음
4. Catalog 종료 후 생성: 503, 기존 목록 단독 조회는 계속 가능
5. 읽음 상태를 true/false로 반복 변경: 모두 성공
6. 대시보드 조회 중 한 서버 실패: 다른 조회에 취소 신호 전달
7. 서버 포트를 다른 프로세스가 사용 중: Fx startup 실패
8. MySQL 중단: 원본 SQL 오류 대신 일반적인 외부 오류

분리된 서버 모듈의 검증은 위 명령으로 실행한다. 기초 예제와 worker pool 테스트는 go-basic에서 별도로 실행한다.

## 다음 실무 과제

- 의존성 readiness, 요청 ID, tracing과 metrics, DB pool 통계 추가
- 목록 소유권과 인증 주체를 전달하고 권한 검사
- 페이지네이션·필터를 추가한 뒤 생성 SQL과 EXPLAIN 확인
- 사용자 입력을 기록하는 로깅 정책과 요청별 로그 상관관계 설계
- 실제 운영 비밀 관리, TLS, 배포 전용 마이그레이션 단계 구성
- 같은 키 재시도에서 Catalog 장애 의존성을 줄이도록 조회 순서 개선
- Catalog 자료 URL 검사 배치를 bounded worker pool로 구현
- GORM Repository를 sqlc 구현으로 교체하고 유지보수 비용 비교
