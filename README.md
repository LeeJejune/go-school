# go-school — Java/Spring 개발자의 Go 입문

Java/Spring으로 백엔드를 만들어 본 사람이 Go를 처음 배운다고 가정한 학습 저장소입니다. **문법을 작은 프로그램으로 익히는 단계**와 **서버 요청을 끝까지 따라가는 단계**를 나눴습니다.

| 프로젝트 | 처음 하는 일 | 준비 |
| --- | --- | --- |
| [go-basic](go-basic/README.md) | 실행·변수·함수부터 다섯 단계로 학습 | Go 1.26 이상 |
| [go-practice](go-practice/README.md) | Spring의 Controller·Service·Repository에 대응하는 코드 찾기 | Go 1.26 이상, Docker |

## 처음이라면 여기서 시작하세요

저장소 루트인 go-school에서 실행합니다.

```bash
make -C go-basic basics
```

그다음 [기초 1장](go-basic/docs/01-basics.md)을 읽으세요. 문서에 있는 작은 전체 예제와 예상 출력을 확인하고, 문제 하나씩 바꿔 보는 방식입니다.

기초 1~4장을 읽으면 서버 실습을 시작할 수 있습니다. 동시성 장의 worker pool 내부 구현은 나중에 읽어도 됩니다.

## 서버를 실행하고 싶다면

Docker Desktop을 켠 뒤 저장소 루트에서 실행합니다.

```bash
make -C go-practice up
```

http://localhost:8080/v1/resources 에서 JSON 응답을 확인합니다. 처음에는 빈 배열이 정상입니다. [HTTP 요청 파일](go-practice/requests/school.http)이나 [Postman 컬렉션](go-practice/requests/go-practice.postman_collection.json)으로 자료를 등록하세요. 목록 생성과 읽음 상태 변경은 [서버 README](go-practice/README.md), 실제 코드를 읽는 시작점은 [Spring 역할 대응 문서](go-practice/docs/01-spring-to-go.md)에 있습니다.

종료할 때는 다음을 실행합니다. DB 볼륨은 유지합니다.

```bash
make -C go-practice down
```

## 공통 검사

```bash
make check
make test
make fmt
```

각 프로젝트는 독립적인 go.mod와 Makefile을 갖습니다. 루트의 go.work는 함께 탐색하기 위한 설정이고, 프로젝트별 Makefile은 GOWORK=off로 독립 실행합니다.

실제 DB 테스트는 다음처럼 준비한 뒤 실행합니다.

```bash
make -C go-practice db
make -C go-practice migrate
make -C go-practice test-db
```

VS Code 디버그 설정은 각 프로젝트 폴더의 .vscode에 있습니다. 변경은 [커밋 규칙](CONTRIBUTING.md)에 따라 기능·설정 단위로 기록합니다.
