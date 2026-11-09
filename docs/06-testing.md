# 06. 테스트와 다음 단계

## 준비된 테스트

| 위치 | 확인할 동작 |
| --- | --- |
| internal/reading/service_test.go | 입력 중복 거절, 제목 정규화, 외부 조회 실패 시 저장 금지, 오류 매핑 |
| internal/workpool/map_test.go | 입력 순서 보존, 동시 작업 수 상한, 실패와 취소 전파 |
| internal/mysqlstore/store_integration_test.go | 동시 동일 키 생성, 내용 충돌, 트랜잭션 롤백, 읽음 반복 설정 |

테스트 코드는 포함했지만 이번 작업에서는 실행하지 않았다. MySQL 테스트는 `MYSQL_TEST_DSN`이 없으면 skip한다. 단위 테스트 통과와 실제 SQL 통합 테스트 통과를 구분해야 한다.

```bash
make check
make db
make test-db
```

DB 테스트는 별도 로컬 학습 DB에서 실행한다. 테이블을 TRUNCATE하지 않고 테스트가 만든 키의 행만 지운다. 운영 DB의 DSN을 전달하지 않는다. 초기 스키마가 적용돼 있어야 한다.

## 수동으로 확인할 실패 경로

1. 같은 requestKey로 같은 요청을 2번: 목록 ID 동일
2. 같은 requestKey에 제목 변경: 409
3. 없는 자료 선택: 400, DB에 목록 없음
4. Catalog 종료 후 생성: 503, 저장된 목록 조회는 계속 가능
5. 읽음 상태를 두 번 true로 변경: 모두 성공
6. MySQL 중단: 원본 SQL 정보 대신 외부에는 일반 오류

Gateway부터 MySQL까지의 실제 네트워크 연동과 컨테이너 기동은 별도로 확인할 항목이다.

## 다음 과제

**기초**: 읽음/미읽음 필터 추가. 페이지 크기를 제한하고 조건별 EXPLAIN 비교. 요청이 취소되면 SQL과 하위 RPC까지 중단되는지 확인.

**동시성**: 여러 자료 조회를 worker pool로 병렬 처리해 본 뒤 현재 batch 조회와 RPC 횟수, 지연, 실패 처리 비교. 학습 때문에 더 복잡하게 만든 구조가 실제로도 유리한지 측정.

**운영**: 요청 ID, gRPC interceptor, DB pool 통계, readiness 추가. 인증 주체와 목록 소유권 검증. 마이그레이션 이력 관리 도구 도입.

**설계**: Catalog가 중단돼도 기존 requestKey의 재요청을 처리하도록 조회 순서 변경. 저장 성공 직후 응답이 유실되는 상황을 재현하고 멱등성이 지켜지는지 확인.

**성능**: 현재 20개 제한에서 불필요한 최적화를 하기보다 항목 수와 동시 요청 수를 늘린 뒤 benchmark와 pprof로 병목 확인. throughput과 평균만 보지 말고 지연 분포, 오류와 DB 대기도 확인.
