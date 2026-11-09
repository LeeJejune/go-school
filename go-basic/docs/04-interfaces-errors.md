# 04. implements와 try/catch 없이 읽는 Go 코드

이 장의 목표는 [기본 예제](../examples/01-basics/main.go)의 printLabel과 NewArticle을 이해하는 것입니다. 앞 장의 Article, 메서드, 포인터를 알고 있다고 가정합니다.

모든 명령은 go-basic 폴더에서 실행합니다. package main으로 시작하는 예시는 scratch/main.go에 넣어 실행할 수 있습니다.

## 1. interface는 필요한 메서드를 적은 약속이에요

Java에서는 보통 class Article implements Labeler라고 선언하고 메서드를 구현합니다. Go는 필요한 메서드가 있으면 인터페이스를 만족합니다.

아래는 독립 실행용 전체 예시입니다.

```go
package main

import "fmt"

type Labeler interface {
    Label() string
}

type Article struct {
    Title string
}

func (a Article) Label() string {
    return a.Title
}

func printLabel(value Labeler) {
    fmt.Println(value.Label())
}

func main() {
    article := Article{Title: "Go 공부"}
    printLabel(article)
}
```

```text
Go 공부
```

Labeler는 Label() string 메서드가 필요하다고 적습니다. Article이 그 메서드를 제공하므로 printLabel에 전달할 수 있습니다. Label을 다른 이름으로 바꾸면 컴파일 시 인터페이스를 만족하지 않는다는 오류가 납니다.

여기서 인터페이스를 구현한다는 뜻은 클래스 상속이나 Spring Bean 등록과 관계없습니다. 메서드의 이름과 타입이 약속을 만족하는지 확인하는 것입니다.

## 2. 실제 예제는 fmt.Stringer를 써요

표준 라이브러리의 fmt.Stringer는 다음 약속을 갖습니다.

```go
type Stringer interface {
    String() string
}
```

저장소의 Article에는 String() string 메서드가 있고, printLabel은 fmt.Stringer를 받습니다. 위 Labeler 예시와 같은 원리입니다.

```go
func (a Article) String() string {
    return fmt.Sprintf("%s (read=%t)", a.Title, a.Read)
}
```

fmt.Sprintf는 문자열을 출력하지 않고 만들어 반환합니다. %s 자리에는 문자열을, %t 자리에는 true/false를 넣습니다. Java에서 toString()을 구현하던 목적과 연결해서 보면 됩니다.

## 3. 실패할 수 있는 함수는 결과와 error를 돌려줘요

Java에서 입력이 잘못되면 예외를 던지던 자리를 Go에서는 다음처럼 작성할 수 있습니다.

```go
package main

import (
    "errors"
    "fmt"
    "strings"
)

func createTitle(input string) (string, error) {
    title := strings.TrimSpace(input)
    if title == "" {
        return "", errors.New("제목을 입력하세요")
    }
    return title, nil
}

func main() {
    title, err := createTitle("   ")
    if err != nil {
        fmt.Println("실패:", err)
        return
    }
    fmt.Println("성공:", title)
}
```

```text
실패: 제목을 입력하세요
```

(string, error)는 반환값이 두 개라는 뜻입니다. 호출한 쪽도 title, err처럼 두 변수로 받습니다.

| 반환 | 의미 |
| --- | --- |
| return title, nil | 성공한 제목과 '오류 없음' |
| return "", err | 이 호출에서는 제목 대신 오류를 처리해야 함 |

nil은 여기서 오류가 없다는 뜻으로 사용했습니다. Go에서 nil을 사용할 수 있는 타입은 정해져 있으며, int나 일반 구조체의 기본값을 모두 nil이라고 부르지는 않습니다.

err != nil이면 실패를 처리하고 보통 return으로 그 경로를 끝냅니다. 정상 흐름은 그 아래로 계속 내려갑니다.

입력을 "  Go 공부  "로 바꿔 다시 실행하면 성공: Go 공부가 나옵니다.

## 4. NewArticle은 생성자 역할을 하는 일반 함수예요

실제 함수의 선언을 읽어 보세요.

```go
func NewArticle(title string) (Article, error)
```

'제목을 받아 Article 값과 error 값을 함께 반환하는 함수'입니다. New로 시작하는 이름은 생성 함수에 자주 쓰는 관례입니다. Spring이 이름을 보고 자동으로 호출하는 함수는 아닙니다.

함수는 빈 제목일 때 Article{}와 오류를 반환합니다. Article{}는 필드가 기본값인 구조체 값입니다. 따라서 반환받은 Article을 바로 쓰기 전에 err를 확인해야 합니다.

예제의 main은 고정된 정상 제목을 넣고, 예상 밖의 실패에는 panic을 사용합니다. 서버에서 잘못된 사용자 입력을 처리하는 방법은 [go-practice의 오류 처리](../../go-practice/docs/02-grpc.md)에서 이어집니다.

## 5. 오류를 감싸도 원인을 확인할 수 있어요

[실제 코드](../examples/01-basics/main.go)는 errEmptyTitle을 만들고 다음처럼 감쌉니다.

```go
return Article{}, fmt.Errorf("new article: %w", errEmptyTitle)
```

new article이라는 설명을 덧붙이면서 원래 오류도 유지하는 방식입니다. errors.Is(err, errEmptyTitle)로 이 원인인지 확인할 수 있습니다. 오류 메시지 문자열 전체를 비교하지 않아도 됩니다.

make basics의 마지막 줄인 empty title: true는 빈 제목으로 생성했을 때 그 오류를 확인했다는 뜻입니다.

## 6. defer는 함수가 끝날 때 정리할 일을 예약해요

```go
package main

import "fmt"

func main() {
    defer fmt.Println("정리")
    fmt.Println("작업")
}
```

```text
작업
정리
```

Java의 finally처럼 자원 정리를 보장하려는 목적에 사용할 수 있습니다. Go의 defer는 현재 함수가 반환되기 직전에 실행됩니다. if 블록을 빠져나올 때마다 실행되는 것은 아닙니다. 여러 개를 예약하면 나중에 예약한 것부터 실행합니다.

## 직접 확인해 보세요

1. implements 없이도 Article을 Labeler로 전달할 수 있는 이유는 무엇인가요?
2. createTitle이 반환한 err가 nil이면 무엇을 뜻하나요?
3. defer로 예약한 출력은 fmt.Println("작업")보다 먼저 실행되나요?

[이전: 구조체·메서드](03-struct-methods.md) · [다음: 동시성](05-concurrency.md)

더 읽기: [Go 공식 오류 처리 소개](https://go.dev/blog/error-handling-and-go), [fmt.Stringer](https://pkg.go.dev/fmt#Stringer).
