# 02. 동시성과 취소

## goroutine이란

`go fn()`은 함수를 새 goroutine에서 실행하게 한다. goroutine은 Go 런타임이 관리하는 실행 흐름이다. OS thread와 1:1로 만드는 구조가 아니며 런타임이 여러 goroutine을 thread들에 배치한다. 시작 비용이 상대적으로 작아도 메모리, 연결과 CPU는 유한하다.

동시성은 여러 작업의 진행을 겹쳐 다루는 구조다. 병렬성은 실제로 같은 순간 여러 작업을 실행하는 것이다. 한 코어에서도 I/O 대기와 다른 작업을 겹칠 수 있지만 CPU 연산을 goroutine으로 쪼갰다고 무조건 빨라지지 않는다.

`main`이 끝나면 남은 goroutine이 완료되기를 자동으로 기다리지 않는다. `sync.WaitGroup`으로 작업 수와 완료를 맞춘다. WaitGroup은 기다리는 도구이지 공유 변수 보호 도구가 아니다.

## channel

channel은 타입이 있는 값 전달 통로다. `make(chan int)`는 버퍼가 없어서 상대방이 받을 준비가 될 때 전송이 진행된다. 버퍼가 있으면 빈 공간까지 보낼 수 있지만 가득 차면 기다린다. 무한 큐가 아니다.

```go
select {
case jobs <- value:
case <-ctx.Done():
    return ctx.Err()
}
```

받는 쪽이 사라져도 전송자는 저절로 종료되지 않는다. 취소 경로가 없으면 영원히 기다리는 goroutine이 남는다. `select`로 전송/수신과 취소를 함께 기다린다. 여러 case가 준비됐을 때 취소가 항상 최우선인 것은 아니다.

전송이 끝났다는 것을 아는 쪽이 channel을 닫는다. 닫힌 channel을 range로 읽으면 남은 버퍼를 소비한 뒤 끝난다. 닫은 뒤 전송하거나 두 번 닫으면 panic이다. nil channel의 전송과 수신은 계속 기다리므로 의도하지 않은 nil도 확인한다.

## Mutex와 데이터 레이스

여러 goroutine이 같은 map을 수정하거나 같은 카운터를 읽고 바꾸면 접근을 보호해야 한다. `sync.Mutex`로 짧은 임계 구역을 감싸거나 단일 소유자에게 channel로 작업을 전달한다. channel이 있다는 사실만으로 다른 공유 메모리도 안전해지지는 않는다.

데이터 레이스는 동기화되지 않은 메모리 접근 문제다. `go test -race`로 실행 경로에서 탐지할 수 있다. 검사에 통과해도 서로 다른 두 요청이 같은 보상을 발급하는 업무상 경쟁 조건이 사라졌다고 보장할 수는 없다. DB 제약과 트랜잭션이 필요한 문제도 있다.

## Context

Context는 요청의 deadline과 취소 신호를 아래 호출로 전달한다. HTTP 요청이 취소되면 RPC와 DB가 불필요한 일을 멈출 수 있도록 같은 흐름의 ctx를 전달한다. 함수의 첫 인자로 받고 보통 struct 필드에 장기간 저장하지 않는다.

```go
ctx, cancel := context.WithTimeout(parent, 300*time.Millisecond)
defer cancel()
response, err := client.BatchGetResources(ctx, request)
```

cancel은 함수를 강제로 죽이는 명령이 아니다. 작업이 ctx를 확인하거나 `QueryContext` 같은 취소 가능한 API를 사용해야 반응한다. 메모리 연산을 무한 반복하는 함수가 ctx를 보지 않으면 종료되지 않는다. 하위 호출의 유효 deadline은 상위에 남은 시간보다 길어지지 않는다.

## worker pool 코드 읽기

[`Map`](../internal/workpool/map.go)은 인덱스를 jobs channel에 보내고, 제한된 worker가 값을 처리한다. 결과는 각 입력의 인덱스에 저장하므로 완료 순서가 달라도 출력 순서는 유지된다.

1. producer 하나가 입력 인덱스를 보내고 jobs를 닫는다
2. worker는 jobs 또는 취소 신호를 기다린다
3. 첫 오류는 sync.Once로 보관하고 다른 작업을 취소한다
4. producer와 worker가 모두 끝난 뒤 결과를 반환한다

서로 다른 slice 항목을 한 worker씩만 쓰는 설계다. append로 같은 slice 길이를 동시에 바꾸는 방식과 다르다. fn 안에서 별도의 공유 상태를 수정한다면 그 상태도 따로 보호해야 한다.

동시 작업 수는 실제 리소스와 작업 특성을 보고 제한한다. 작업이 많다는 이유로 무한히 goroutine을 만들면 대기와 메모리 사용량이 늘 수 있다.

## 직접 바꿔볼 것

다음 단계인 go-practice에서 서버의 사용 예는 [`gateway/dashboard.go`](../../go-practice/internal/gateway/dashboard.go)다. GET /v1/dashboard/{id}가 Catalog와 Reading을 errgroup으로 동시에 호출한다. 둘 중 하나가 실패하면 다른 작업에 취소를 전달하고 둘 다 종료되기를 기다린다. worker pool 예제는 별도 학습 코드이며 이 RPC 경로에서 사용하지 않는다.

- worker 수를 1, 2, 4로 바꾸고 I/O 대기형 작업과 CPU 작업 비교
- 작업 중 하나에 오류를 반환하고 producer까지 종료되는지 확인
- ctx를 무시하는 작업을 넣고 취소의 한계 설명
- `select`에서 취소 경로를 지웠을 때 어떤 경우 전송이 막히는지 생각하기

읽기: [Pipelines and cancellation](https://go.dev/blog/pipelines), [Context](https://go.dev/blog/context), [Race Detector](https://go.dev/doc/articles/race_detector).
