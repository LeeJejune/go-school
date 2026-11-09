# go-school

Go 언어 자체를 익히는 **go-basic**과 실무형 백엔드를 만드는 **go-practice**를 나눈 학습 저장소다.

| 프로젝트 | 배우는 내용 | 준비 |
| --- | --- | --- |
| [go-basic](go-basic/README.md) | 문법, struct/interface/error, goroutine/channel/Context, worker pool | Go 1.26 이상 |
| [go-practice](go-practice/README.md) | Kratos, Fx DI, GORM, Goose, MySQL, gRPC, Gateway | Go 1.26 이상, Docker |

각 디렉터리에 독립적인 go.mod, README와 Makefile이 있다. go-basic은 표준 라이브러리만 사용하며 서버 프레임워크나 MySQL을 설치하지 않아도 된다. 루트의 go.work는 두 프로젝트를 편집기에서 함께 탐색하기 위한 설정이다. 각 프로젝트의 Makefile은 GOWORK=off로 독립 실행을 검증한다.

## 기초 실습

```bash
cd go-basic
make basics
make concurrency
make check
```

어떤 파일부터 읽을지는 [기초 실습 안내](go-basic/README.md)를 본다.

## 서버 실습

저장소 루트에서 다음 명령을 실행한다.

```bash
make -C go-practice up
```

브라우저나 curl로 http://localhost:8080/v1/resources 를 확인한다. 로컬 디버깅, API 사용법, Spring과의 비교는 [서버 실습 안내](go-practice/README.md)에 정리했다.

```bash
make -C go-practice down
```

## 공통 명령

```bash
make check    # 두 프로젝트의 vet + race 검사
make test     # 두 프로젝트의 단위 테스트
make fmt
```

DB 통합 테스트는 서버 프로젝트에서 따로 준비하고 실행한다.

```bash
make -C go-practice db
make -C go-practice migrate
make -C go-practice test-db
```

VS Code 디버그 설정은 각 프로젝트의 .vscode에 있으므로 원하는 프로젝트 폴더를 열어 사용한다. 변경은 [커밋 규칙](CONTRIBUTING.md)에 따라 기능·설정 단위로 기록한다.
