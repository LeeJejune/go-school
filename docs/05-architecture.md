# 05. 클린 아키텍처가 필요한가

## 이 저장소의 선택

기능 기준 패키지를 두고 DB와 외부 RPC 경계만 작게 나눴다. Catalog는 단순한 읽기 기능이어서 별도 repository를 만들지 않았다. Reading은 입력 검증과 외부 조회, 트랜잭션 저장이 이어져 저장소와 클라이언트를 분리했다.

```text
cmd/reading                실제 객체 생성과 연결
    ├── reading            업무 흐름, 모델, 필요한 interface
    ├── mysqlstore         SQL 구현 → reading의 모델 사용
    └── catalogclient      gRPC 구현 → reading의 모델 사용

reading                    mysqlstore와 catalogclient를 import하지 않음
```

Reading 서비스에는 gRPC 요청/응답 변환도 함께 있다. 완전히 transport와 독립된 domain/application 구조는 아니다. 전송 계층까지 분리해야 할 실제 필요가 아직 작아서 한 단계 줄였다. HTTP와 메시지 소비자가 같은 업무 로직을 호출하게 된다면 application 함수를 분리할 지점이다.

## 이름보다 의존성 방향

클린/육각형 아키텍처의 핵심은 폴더 이름보다 업무 규칙이 구체적인 DB와 외부 API에 묶이지 않게 하는 방향이다. `reading.Store`를 MySQL이 구현하고 main이 전달한다. 서비스가 MySQL 드라이버를 생성하거나 SQL을 알 필요가 줄어든다.

테스트하기 편해지는 이점은 있지만 interface와 모델 변환을 유지하는 비용도 생긴다. CRUD 하나에 controller, usecase, service, domain, repository를 모두 만들고 값 전달만 반복하면 변경 위치만 많아질 수 있다.

## 어느 정도로 나눌까

| 상황 | 출발점 | 더 나눌 신호 |
| --- | --- | --- |
| 작은 읽기 전용 API | 패키지 하나와 구체 타입 | 외부 의존성 교체나 테스트 격리가 반복 필요 |
| 검증과 저장 흐름이 있는 기능 | 기능 패키지 + 좁은 저장소 interface | 여러 진입점에서 같은 업무 규칙 호출 |
| 업무 규칙이 복잡한 제품 | application/domain/adapter 경계 검토 | 인프라 변경이 업무 테스트까지 흔듦 |
| 여러 팀의 독립 서비스 | 명확한 소유권과 계약 | 독립 배포/스케일/장애 격리가 운영 비용을 정당화 |

분리 자체가 목적이면 네트워크 실패와 배포 복잡도만 늘어난다. 학습용 서버 분리를 실제 제품의 필수 구조로 일반화하지 않는다.

## Go에서 interface를 둘 위치

호출하는 쪽이 필요한 행동을 정의한다. `Store`는 목록 서비스가 필요한 동작이지 모든 DB에 통용되는 CRUD interface가 아니다. 저장소마다 Create, Update, Delete 메서드를 기계적으로 맞추지 않는다. 실제 두 구현이 없더라도 외부 의존성을 격리할 이유가 있으면 interface가 유용할 수 있지만, 테스트를 위해 모든 함수를 추상화할 필요는 없다.

## 직접 비교할 변경

1. 간단한 Catalog에서 interface를 추가한다면 이점이 실제로 있는지 설명
2. Reading의 gRPC 변환과 업무 함수 분리: 재사용 이점과 추가 코드 비교
3. `Store`를 mock으로만 테스트할 때 SQL 오류를 놓치는 이유 찾기
4. 다른 서비스에 자료 삭제 기능이 생기면 snapshot, 최신 조회, 동기화 중 무엇을 고를지 결정

읽기: [Go 공식 모듈 구조](https://go.dev/doc/modules/layout), [Three Dots Labs의 실제 Go 리팩터링 사례](https://threedots.tech/post/introducing-clean-architecture/), [당근의 계층형과 육각형 아키텍처 사례](https://medium.com/daangn/%EC%95%84%ED%82%A4%ED%85%8D%EC%B2%98%EC%97%90-%EB%8C%80%ED%95%9C-%EA%B3%A0%EB%AF%BC%EC%9D%80-%EC%B2%98%EC%9D%8C%EC%9D%B4%EB%9D%BC-b75dffd73eb0).

공식 Go 구조 문서는 언어 도구의 경계와 관례를 설명하고, 두 회사의 글은 저자들이 선택한 사례다. 하나의 표준 설계로 취급하지 않고 문제와 규모를 비교한다.
