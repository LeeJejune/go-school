# 01. Java의 main 메서드에서 Go의 main 함수로

이 장의 목표는 **Go 프로그램을 실행하고, 변수와 함수가 적힌 코드를 읽는 것**입니다. Spring 설정이나 서버는 아직 다루지 않습니다.

## 1. 먼저 기존 예제를 실행하세요

아래 명령은 저장소의 go-basic 폴더에서 실행합니다.

```bash
make basics
```

다음 세 줄이 나오면 준비가 끝났습니다.

```text
A Tour of Go (read=true)
read=1 unread=1
empty title: true
```

[실행한 파일](../examples/01-basics/main.go)을 열고 맨 아래의 func main을 찾으세요. 지금은 나머지 함수를 모두 이해하지 않아도 됩니다.

## 2. Java와 비교할 가장 작은 프로그램

Java에서 public static void main으로 시작하던 프로그램을 Go에서는 다음처럼 작성합니다. 아래는 파일 전체를 실행할 수 있는 예시입니다.

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go")
}
```

```text
Hello, Go
```

| 코드 | 지금 알아둘 의미 |
| --- | --- |
| package main | 실행할 프로그램이 속한 패키지입니다. |
| import "fmt" | 문자열 출력 등을 제공하는 표준 패키지를 가져옵니다. |
| func main() | 프로그램의 시작 함수입니다. 클래스 안에 넣지 않습니다. |
| fmt.Println(...) | Java의 System.out.println처럼 값을 출력합니다. |

Java 문장 끝에 붙이던 세미콜론은 일반적인 Go 코드에서 생략합니다. if나 함수의 여는 중괄호는 같은 줄에 씁니다.

이후 문서에서 **package main으로 시작하는 예시**는 독립 실행용 전체 코드입니다. 따라 치려면 go-basic 아래 scratch 폴더를 만들고 scratch/main.go에 예시 하나를 넣은 뒤 실행하세요.

```bash
go run ./scratch
```

다음 예시를 실행할 때는 같은 파일의 내용을 교체하세요. package main이 없는 짧은 코드는 설명에 필요한 일부이므로 그것만 붙여 넣어 실행하지 않습니다.

## 3. 변수는 '이름 타입' 순서로 적어요

Java의 String title에 대응하는 모양은 Go의 title string입니다.

```go
package main

import "fmt"

func main() {
    var title string = "Go 공부"
    count := 2
    count = count + 1

    fmt.Println(title)
    fmt.Println(count)
}
```

```text
Go 공부
3
```

- var title string은 타입을 적어 변수를 선언합니다.
- count := 2는 함수 안에서 변수를 새로 만들고 타입을 추론합니다. 여기서는 int입니다.
- 이미 있는 변수에 다시 값을 넣을 때는 =를 씁니다.

Java의 var처럼 타입 추론이 있어도 동적 타입 언어가 되는 것은 아닙니다. count에 나중에 문자열을 넣으면 컴파일 오류가 납니다.

초기값을 적지 않은 Go 변수에는 기본값인 **제로 값**이 들어갑니다. int는 0, bool은 false, string은 빈 문자열입니다. Java의 지역 변수와 달리 var count int만 선언해도 0으로 읽을 수 있습니다.

## 4. 함수도 타입이 뒤에 와요

다음은 함수와 조건문, 반복문을 함께 실행하는 전체 예시입니다.

```go
package main

import "fmt"

func add(a int, b int) int {
    return a + b
}

func main() {
    total := 0
    for i := 1; i <= 3; i++ {
        total = add(total, i)
    }

    if total == 6 {
        fmt.Println("합계:", total)
    }
}
```

```text
합계: 6
```

func add(a int, b int) int를 왼쪽부터 읽으면 '이름은 add, 정수 두 개를 받아 정수 하나를 반환하는 함수'입니다. Java라면 int add(int a, int b)에 가깝습니다.

if의 조건에는 괄호가 필요하지 않습니다. 반복에는 for를 사용합니다. Go에서 while처럼 반복하고 싶을 때도 조건만 적은 for를 씁니다.

## 5. go.mod와 package는 다른 말이에요

[go.mod](../go.mod)는 이 프로젝트의 모듈 이름과 Go 버전을 적는 파일입니다. 의존성을 관리하는 역할은 Java의 build.gradle이나 pom.xml을 떠올리면 됩니다. 같은 디렉터리의 Go 파일들은 보통 같은 package에 속합니다.

이 예제는 다음 두 개의 실행 패키지를 갖습니다.

```text
examples/01-basics/       기본 문법 프로그램
examples/02-concurrency/  동시성 프로그램
```

make basics는 Makefile에 적힌 go run ./examples/01-basics를 대신 실행합니다. Go가 Makefile을 반드시 요구하는 것은 아닙니다.

가져오고 쓰지 않은 import나 사용하지 않은 지역 변수 때문에 컴파일 오류가 날 수 있습니다. 연습 중에는 메시지가 가리키는 줄부터 확인하세요.

## 직접 확인해 보세요

1. 첫 Hello 예시의 문구를 바꿔 실행하세요.
2. add가 계산한 합계가 10이 되게 반복 범위를 바꿔 보세요.
3. count := 2 다음 줄을 count = "둘"로 바꾸면 어떤 오류가 나는지 확인하세요.

다음 장으로 넘어가기 전, func main과 :=, =의 역할을 설명할 수 있으면 충분합니다.

[다음: 배열·slice·map](02-collections.md) · [학습 순서](../README.md)

더 읽기: [A Tour of Go의 패키지와 함수](https://go.dev/tour/basics/1).
