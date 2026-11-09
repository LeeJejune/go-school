# go-basic — Java 개발자의 첫 Go 실습

Java/Spring으로 백엔드를 개발해 봤지만 Go 문법은 처음인 사람을 위한 프로젝트입니다. 여기서는 Go와 표준 라이브러리만 사용합니다. Docker나 DB는 준비하지 않아도 됩니다.

## 1. 먼저 실행 결과 하나를 보세요

편집기에서 go-basic 폴더를 열고, 터미널도 그 폴더에서 실행하세요. 저장소 루트인 go-school에 있다면 먼저 이동합니다.

```bash
cd go-basic
go version
make basics
```

Go 1.26 이상이 필요합니다. 다음 결과가 나오면 시작할 준비가 됐습니다.

```text
A Tour of Go (read=true)
read=1 unread=1
empty title: true
```

마지막 true는 오류가 없다는 뜻이 아닙니다. 빈 제목으로 생성했을 때 **기대한 오류가 맞는지 확인한 결과**입니다. 자세한 설명은 4장에서 읽습니다.

## 2. 문서는 아래 순서로 읽으세요

| 순서 | 문서 | 이 장을 읽고 할 수 있는 것 |
| --- | --- | --- |
| 1 | [실행·변수·함수](docs/01-basics.md) | func main, :=, if, for를 읽습니다. |
| 2 | [배열·slice·map](docs/02-collections.md) | List와 Map에 해당하는 자료를 다룹니다. |
| 3 | [구조체·메서드·포인터](docs/03-struct-methods.md) | Article과 *Article의 차이를 설명합니다. |
| 4 | [인터페이스·오류·defer](docs/04-interfaces-errors.md) | implements 없는 구현과 (값, error)를 읽습니다. |
| 5 | [goroutine·channel·취소](docs/05-concurrency.md) | 작업을 따로 실행하고 완료를 기다립니다. |

1~4장을 읽은 다음 [기존 예제의 main](examples/01-basics/main.go)을 다시 따라가세요. NewArticle → MarkRead → printLabel → slice와 map 순서가 읽히면 됩니다.

5장의 worker pool 내부 구현은 선택 학습입니다. 처음부터 제네릭과 sync.Once를 모두 이해하려고 하기보다, 작은 goroutine 예제를 먼저 실행하세요.

## 3. 설명 예제를 따로 실행하는 방법

문서에서 package main으로 시작하는 코드는 독립적으로 실행할 수 있는 전체 예시입니다.

1. go-basic 아래 scratch 폴더를 만듭니다.
2. scratch/main.go 파일을 만들고 예시 하나를 넣습니다.
3. go-basic 터미널에서 go run ./scratch를 실행합니다.
4. 다음 예시를 볼 때 파일 내용을 교체합니다.

package main이 없는 짧은 코드는 설명용 일부입니다. 설명과 함께 실제 저장소 파일에서 찾아보세요. 학습을 마친 scratch 폴더는 연습용이므로 커밋할 필요가 없습니다.

## 4. 이미 있는 프로그램과 테스트

| 명령 | 실행하는 것 |
| --- | --- |
| make basics | examples/01-basics/main.go |
| make concurrency | examples/02-concurrency/main.go |
| make test | 내부 패키지의 테스트 |
| make check | go vet + race detector를 켠 테스트 |

코드를 바꿔도 이전 실행 결과가 자동으로 바뀌지는 않습니다. 명령을 다시 실행하세요. 문서가 제안한 수정 실험은 끝나면 원래대로 복구합니다.

VS Code에서는 이 폴더의 .vscode/launch.json을 사용해 Basics 또는 Concurrency를 선택하여 디버깅할 수 있습니다. go.mod는 이 프로젝트의 이름과 Go 버전을 관리합니다. 외부 모듈 의존성이 없어 go.sum은 두지 않습니다.

## 5. 다음 단계

기본 문법이 읽히면 [go-practice](../go-practice/README.md)로 넘어가세요. 익숙한 Controller·Service·Repository·Entity가 Go 코드에서 어디에 있는지부터 설명합니다.

[저장소 전체 안내](../README.md) · [커밋 규칙](../CONTRIBUTING.md)
