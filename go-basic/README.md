# go-basic

Go 언어와 표준 라이브러리만 사용하는 기초 실습 프로젝트다. DB, Docker, 서버 프레임워크는 필요 없다.

## 실행

go-basic 폴더에서 명령을 실행한다.

```bash
make basics
make concurrency
make check
```

Go 1.26 이상이 필요하다. 외부 모듈 의존성이 없어 go.sum은 두지 않는다.

## 읽고 바꿔볼 순서

| 순서 | 문서 | 코드 | 실습 |
| --- | --- | --- | --- |
| 1 | [기본 문법](docs/01-basics.md) | [examples/01-basics/main.go](examples/01-basics/main.go) | MarkRead 호출을 빼고 결과 비교 |
| 2 | [동시성과 취소](docs/02-concurrency.md) | [examples/02-concurrency/main.go](examples/02-concurrency/main.go) | worker 수를 1, 2, 4로 변경 |
| 3 | 동시 작업 구현 | [internal/workpool/map.go](internal/workpool/map.go) | 작업 실패와 취소 경로 따라가기 |
| 4 | 동작 검증 | [map_test.go](internal/workpool/map_test.go) | 순서 보존과 동시 작업 수 제한 확인 |

기본 문법 예제는 func main부터 읽는다. NewArticle은 생성·검증·오류 반환, MarkRead는 포인터 메서드, printLabel은 interface, articles와 counts는 slice와 map을 보여 준다.

```text
A Tour of Go (read=true)
read=1 unread=1
empty title: true
```

동시성 예제는 [1 4 9 16 25]를 출력한다. worker들의 완료 순서와 관계없이 입력 순서로 결과를 돌려준다.

## 프로젝트 구조

```text
examples/
  01-basics/       실행 가능한 문법 예제
  02-concurrency/  실행 가능한 동시성 예제
internal/workpool/ worker pool과 테스트
docs/              언어 설명
go.mod             표준 라이브러리만 사용하는 독립 모듈
```

VS Code에서는 이 폴더를 열고 Go 확장을 설치한다. 실행 및 디버그에서 Basics나 Concurrency를 선택할 수 있다. 코드 변경 후 명령을 다시 실행한다.

다음 단계는 [go-practice](../go-practice/README.md)다. 기초 문법이 서버의 Controller, Service, Repository와 동시 호출에 어떻게 쓰이는지 이어서 확인한다.

공식 자료: [A Tour of Go](https://go.dev/tour/), [Code Review Comments](https://go.dev/wiki/CodeReviewComments), [Context](https://go.dev/blog/context).
