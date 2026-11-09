# 05. goroutine을 작은 예제부터 이해하기

이 장의 목표는 **작업을 따로 시작하는 것과 완료를 기다리는 것을 구분하는 것**입니다. 처음에는 1~3절만 이해해도 됩니다. 마지막 worker pool 내부 구현은 선택 학습입니다.

모든 명령은 go-basic 폴더에서 실행합니다. 전체 예시는 scratch/main.go에 넣고 go run ./scratch로 실행할 수 있습니다.

## 1. Java에서 작업을 따로 실행하던 상황을 떠올려요

Java에서 ExecutorService에 작업을 제출하면 호출한 코드와 작업이 별도로 진행할 수 있습니다. Go에서는 함수 호출 앞에 go를 붙여 새 **goroutine**에서 실행합니다.

```go
doWork()    // 이 호출이 끝나야 다음 줄로 진행
go doWork() // 별도의 goroutine에서 진행
```

goroutine은 Go 런타임이 관리하는 실행 흐름입니다. goroutine 하나마다 OS 스레드를 하나 만드는 것은 아닙니다. 여러 작업을 진행할 수 있는 구조인 '동시성'과, 실제로 같은 순간 CPU에서 실행되는 '병렬성'은 구분합니다.

그리고 main이 끝나면 다른 goroutine의 종료를 자동으로 기다리지 않습니다. 작업을 시작했으면 언제 완료를 기다릴지도 정해야 합니다.

## 2. channel로 결과를 받고 기다려 봐요

아래는 독립 실행용 전체 코드입니다.

```go
package main

import "fmt"

func main() {
    result := make(chan string)

    go func() {
        result <- "작업 끝"
    }()

    message := <-result
    fmt.Println(message)
    fmt.Println("main 끝")
}
```

```text
작업 끝
main 끝
```

새 문법을 한 줄씩 읽어 보세요.

| 표현 | 의미 |
| --- | --- |
| make(chan string) | 문자열을 전달하는 channel 생성 |
| go func() { ... }() | 이름 없는 함수를 새 goroutine에서 실행 |
| result <- "작업 끝" | channel로 값 보내기 |
| message := <-result | channel에서 값을 받아 변수에 저장 |

여기서는 결과를 받을 때까지 main이 기다립니다. 따라서 임의의 time.Sleep을 넣어 작업이 끝나기를 추측할 필요가 없습니다.

기본 channel은 버퍼가 없어서 보내는 쪽과 받는 쪽이 서로 준비될 때 전달됩니다. 받는 쪽이 영원히 오지 않으면 보내는 작업도 계속 기다릴 수 있습니다.

## 3. Context는 '그만해도 된다'는 신호를 전달해요

작업을 시작한 뒤 사용자가 요청을 취소하거나 제한 시간이 끝날 수 있습니다. 그 사실을 아래 작업에 알려 주는 통로가 Context입니다.

```go
package main

import (
    "context"
    "fmt"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    result := make(chan string)

    go func() {
        <-ctx.Done()
        result <- "취소 신호를 확인하고 종료"
    }()

    cancel()
    fmt.Println(<-result)
}
```

```text
취소 신호를 확인하고 종료
```

WithCancel은 Context와 취소 함수인 cancel을 함께 반환합니다. cancel을 호출하면 ctx.Done()을 기다리는 코드가 신호를 받습니다.

cancel은 다른 함수를 강제로 종료시키지 않습니다. 예시의 goroutine이 ctx.Done()을 확인하기 때문에 반응한 것입니다. 긴 작업에서도 신호를 확인하거나 Context를 받는 API를 사용해야 취소에 반응할 수 있습니다.

제한 시간도 정하려면 WithTimeout을 사용합니다. 저장소의 동시성 예제에는 아래 코드가 있습니다.

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
```

'최대 1초의 마감시간을 가진 Context를 만들고, 함수가 끝나면 관련 자원을 정리한다'고 읽으면 됩니다. 마감시간을 deadline이라고 부릅니다. Context를 전달받는 함수의 첫 인자 이름은 보통 ctx입니다.

## 4. 저장소의 동시성 예제를 실행해 봐요

```bash
make concurrency
```

```text
[1 4 9 16 25]
```

[examples/02-concurrency/main.go](../examples/02-concurrency/main.go)의 호출을 먼저 보세요.

```go
results, err := workpool.Map(ctx, []int{1, 2, 3, 4, 5}, 2, ...)
```

위는 호출의 일부만 보여 준 코드입니다. 마지막 ... 자리에는 실제 파일에 있는 작업 함수가 들어갑니다.

- 입력은 1~5입니다.
- 2는 최대 두 작업자가 나눠 처리한다는 뜻입니다.
- 각 작업 함수는 잠깐 기다린 뒤 숫자의 제곱을 반환합니다.
- 모든 작업이 끝나면 입력 순서대로 결과를 돌려줍니다.

worker pool은 작업을 처리하는 goroutine 수를 정해 두고 작업을 나누는 방식입니다. Java의 고정 크기 작업자 풀과 사용 목적을 연결할 수 있습니다. 제곱 계산에 잠깐 대기를 넣은 것은 작업 흐름을 보여 주기 위한 것이며 성능 비교 결과를 뜻하지는 않습니다.

Map은 이 프로젝트에서 만든 함수 이름입니다. 키·값 자료구조인 소문자 map과 구분하세요.

## 5. 내부 구현은 이 순서로 읽으세요 — 선택 학습

[internal/workpool/map.go](../internal/workpool/map.go)에서 다음 이름을 찾습니다.

| 이름 | 맡은 일 |
| --- | --- |
| jobs channel | 처리할 입력의 인덱스를 작업자에게 전달 |
| go func | 작업을 보내는 goroutine과 처리하는 goroutine 시작 |
| sync.WaitGroup | 시작한 작업들이 모두 끝나기를 기다림 |
| select | 할 일이 오거나 취소 신호가 오는 경우를 함께 확인 |
| sync.Once | 여러 실패 중 첫 오류를 한 번만 기록 |
| results[i] | 완료 순서와 관계없이 원래 위치에 결과 저장 |

WaitGroup의 Add는 기다릴 작업 수를 늘리고, Done은 완료를 알리며, Wait는 완료까지 기다립니다. 공유 변수의 동시 접근을 막는 기능은 아닙니다.

Map[T, R any]의 대괄호는 타입 매개변수입니다. Java의 제네릭 메서드처럼 입력 타입 T와 결과 타입 R을 바꿔 쓸 수 있게 합니다. 이 예제에서는 둘 다 int이고 호출할 때 추론됩니다.

처음에는 jobs와 WaitGroup까지만 따라가세요. sync.Once나 테스트의 atomic 연산까지 한 번에 익힐 필요는 없습니다.

## 6. 같이 실행한다고 공유 데이터가 안전해지지는 않아요

두 goroutine이 같은 카운터나 map을 동시에 변경한다면 보호가 필요합니다. Java에서 동기화나 Lock이 필요하던 문제와 비슷합니다. Go에서는 Mutex나 하나의 소유자가 channel로 작업을 받는 방식 등을 쓸 수 있습니다.

```bash
make check
```

이 명령은 정적 검사와 race detector를 켠 테스트를 실행합니다. race detector는 테스트가 실제로 실행한 경로의 잘못된 동시 메모리 접근을 찾는 데 도움을 줍니다. 검사를 통과했다고 모든 동시성 문제가 사라졌다는 뜻은 아닙니다.

## 직접 확인해 보세요

1. 첫 channel 예시에서 값을 받는 줄이 왜 완료 대기 역할을 하나요?
2. cancel을 호출하기만 하면 Context를 확인하지 않는 함수도 강제로 멈추나요?
3. 기존 예제의 worker 수를 1이나 4로 바꿔도 결과 순서는 유지되나요?

추가 실험은 [기존 예제](../examples/02-concurrency/main.go)의 time.Second를 time.Millisecond로 바꾸는 것입니다. 작업 대기보다 제한 시간이 짧아 오류가 발생합니다. 예제 main이 이 오류를 panic으로 보여 주므로 context deadline exceeded를 확인한 뒤 원래 값으로 복구하세요.

서버에서 독립적인 두 API를 동시에 호출하는 예는 [go-practice의 요청 흐름 문서](../../go-practice/docs/02-grpc.md)에서 이어집니다.

[이전: 인터페이스·오류](04-interfaces-errors.md) · [학습 순서](../README.md)

더 읽기: [Go 공식 goroutine 설명](https://go.dev/doc/effective_go#goroutines), [Context 문서](https://pkg.go.dev/context).
