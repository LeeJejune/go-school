# 03. Java 클래스에서 구조체와 메서드로

이 장의 목표는 [기본 예제](../examples/01-basics/main.go)에 있는 Article과 MarkRead를 읽는 것입니다. 특히 **원본 값을 바꾸려면 왜 *Article을 쓰는지** 확인합니다.

모든 명령은 go-basic 폴더에서 실행합니다. 전체 예시는 scratch/main.go에 넣고 go run ./scratch로 실행할 수 있습니다.

## 1. struct는 관련 데이터를 묶어요

Java에서 문서의 제목과 읽음 상태를 클래스로 묶었다면 Go에서는 다음처럼 쓸 수 있습니다.

```go
type Article struct {
    Title string
    Read  bool
}
```

위 코드는 타입 정의 부분입니다. Article은 타입 이름이고 Title과 Read는 필드입니다.

```go
article := Article{Title: "Go 공부"}
```

이렇게 값을 만듭니다. Read를 생략하면 bool의 제로 값인 false가 들어갑니다. Java의 new Article(...)처럼 보이는 기능을 만들 수 있지만, 여기서는 먼저 구조체 값을 직접 만든 것입니다.

Title처럼 첫 글자가 대문자인 이름은 다른 패키지에서도 사용할 수 있는 공개 이름입니다. 소문자로 시작하면 같은 패키지 안에서 사용합니다. Java의 public/private를 그대로 옮기기보다는 '패키지 밖에서 접근 가능한가'로 이해하세요.

## 2. 구조체를 대입하면 값이 복사돼요

```go
package main

import "fmt"

type Article struct {
    Title string
    Read  bool
}

func main() {
    original := Article{Title: "Go 공부"}
    copied := original
    copied.Title = "제목 변경"

    fmt.Println(original.Title)
    fmt.Println(copied.Title)
}
```

```text
Go 공부
제목 변경
```

Java에서 같은 객체를 가리키는 변수 두 개를 만들던 것과 결과가 다릅니다. 여기서는 Article 값이 복사되어 각 변수의 Title 필드가 따로 바뀝니다.

구조체에 slice나 포인터 같은 필드가 있다면 그 필드가 가리키는 데이터까지 독립적으로 복사되지는 않습니다. 현재 Article은 string과 bool만 갖고 있으니 위 결과부터 익히세요.

## 3. 원본을 바꾸는 주소를 전달할 수 있어요

```go
package main

import "fmt"

type Article struct {
    Title string
    Read  bool
}

func markRead(article *Article) {
    article.Read = true
}

func main() {
    article := Article{Title: "Go 공부"}
    markRead(&article)
    fmt.Println(article.Read)
}
```

```text
true
```

두 기호를 나눠 읽어 보세요.

| 표현 | 의미 |
| --- | --- |
| *Article | Article을 가리키는 포인터 타입 |
| &article | 이 article 값의 주소 |

markRead는 원본을 찾아갈 수 있는 주소를 받았으므로 원본의 Read를 true로 바꿀 수 있습니다. 함수 인자로 주소를 전달하는 경우에도 포인터 값 자체는 복사됩니다.

처음부터 메모리 주소 계산을 공부할 필요는 없습니다. **Article을 받으면 구조체 값의 복사본, *Article을 받으면 원본에 접근할 수 있는 포인터를 받는다**고 생각하세요.

## `*`가 값 앞에 있으면 의미가 달라져요

`*Article`은 타입을 쓰는 자리이고, `*p`는 포인터가 가리키는 값을 읽는 표현입니다.

```go
package main

import "fmt"

type Article struct { Title string }

func main() {
    value := Article{Title: "처음 제목"}
    p := &value        // p의 타입은 *Article
    copied := *p       // p가 가리키는 Article 값을 복사
    p.Title = "변경한 제목"

    fmt.Println(value.Title)
    fmt.Println(copied.Title)
}
```

```text
변경한 제목
처음 제목
```

`p.Title`은 원본 필드를 바꿉니다. 포인터가 가리키는 구조체 필드에 접근할 때 `(*p).Title`을 줄여 쓸 수 있습니다. `copied`는 별도 구조체 값입니다. 구조체 안에 slice·map·포인터가 있다면 내부 데이터까지 독립 복사되는 것은 아닙니다.

서버 코드의 `*catalog.Catalog`도 같은 포인터 타입입니다. `catalog.Catalog`는 catalog 패키지의 Catalog 타입이고 `*`가 그 타입을 가리키는 포인터임을 표시합니다.

`&pb.ListResourcesResponse{}`는 pb 패키지의 응답 구조체를 만들고 주소를 얻습니다. `pb`는 import 별칭이며 `&`나 `*`와 관계없는 이름입니다. [서버 코드에서 pb와 포인터 읽기](../../go-practice/docs/01-spring-to-go.md)에서 실제 import를 확인할 수 있습니다.

## 4. 함수를 타입에 붙이면 메서드가 돼요

저장소는 앞의 함수를 다음 메서드로 작성했습니다.

```go
func (a *Article) MarkRead() {
    a.Read = true
}
```

func 뒤의 (a *Article)은 **수신자(receiver)**입니다. 'Article을 가리키는 값을 a라는 이름으로 받아 실행하는 메서드'라는 뜻입니다.

호출은 다음처럼 합니다.

```go
article.MarkRead()
```

Java에서 메서드 안의 this로 필드를 참조하던 위치에 여기서는 a를 씁니다. a는 직접 붙인 변수명이므로 다른 이름도 가능하지만, 같은 타입의 메서드에서는 일관되게 쓰는 편이 읽기 좋습니다.

주소를 얻을 수 있는 변수인 article에는 컴파일러가 포인터 메서드 호출을 편하게 쓸 수 있도록 도와줍니다. 이 호출에서 매번 (&article).MarkRead()라고 적을 필요는 없습니다.

## 5. 실제 예제에서 원본 변경을 확인하세요

go-basic 폴더에서 make basics를 실행하면 첫 줄은 다음과 같습니다.

```text
A Tour of Go (read=true)
```

[main 함수](../examples/01-basics/main.go)의 article.MarkRead()를 주석 처리한 뒤 다시 실행해 보세요.

```text
A Tour of Go (read=false)
read=0 unread=2
empty title: true
```

확인 후 주석을 원래대로 돌려놓으세요.

이번에는 MarkRead의 수신자만 *Article에서 Article로 바꿔 보세요. 메서드가 복사본을 바꾸기 때문에 main의 article.Read는 false로 남습니다. 실험이 끝나면 *Article로 복구합니다.

## 직접 확인해 보세요

1. Article{Title: "Go"}에서 Read 값은 무엇인가요?
2. Article과 *Article은 원본을 변경할 때 어떤 차이가 있나요?
3. (a *Article)에서 a는 타입인가요, 변수명인가요?

String 메서드가 왜 필요한지와 NewArticle이 오류를 함께 반환하는 이유는 다음 장에서 다룹니다.

[이전: 컬렉션](02-collections.md) · [다음: 인터페이스·오류](04-interfaces-errors.md)

더 읽기: [A Tour of Go의 포인터](https://go.dev/tour/moretypes/1), [메서드](https://go.dev/tour/methods/1).
