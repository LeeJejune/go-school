# 02. 배열, slice, map을 Java 컬렉션과 비교하기

이 장의 목표는 **여러 값을 저장하고 반복문으로 읽는 것**입니다. Java의 배열, List, Map을 써 봤다면 용도부터 연결할 수 있습니다.

모든 명령은 go-basic 폴더에서 실행합니다. package main으로 시작하는 예시는 [1장의 실행 방법](01-basics.md)에 따라 scratch/main.go에 하나씩 넣어 실행할 수 있습니다.

## 1. 우선 이 표만 봐도 돼요

| Go 표현 | 용도 | Java에서 떠올릴 것 |
| --- | --- | --- |
| [3]int | 정수 세 개를 담는 고정 길이 배열 | int[] |
| []int | 길이가 달라질 수 있는 자료 모음 | List<Integer>의 사용 목적 |
| map[string]int | 문자열 키로 정수 찾기 | Map<String, Integer> |

사용 목적을 비교한 표입니다. 특히 slice는 Java의 ArrayList와 내부 동작까지 같은 타입은 아닙니다.

Go 배열은 길이도 타입에 포함합니다. [2]int와 [3]int는 다른 타입입니다. 처음에는 자주 보게 될 slice와 map부터 익히세요.

## 2. slice: 자료 모음을 만들고 추가해요

```go
package main

import "fmt"

func main() {
    titles := []string{"Go", "Spring"}
    titles = append(titles, "MySQL")

    fmt.Println(titles)
    fmt.Println(len(titles))

    for index, title := range titles {
        fmt.Println(index, title)
    }
}
```

```text
[Go Spring MySQL]
3
0 Go
1 Spring
2 MySQL
```

[]string은 '문자열을 담는 slice'입니다. {} 안에는 시작할 때 넣을 값을 적습니다.

append는 항목을 추가한 slice를 반환하므로 반환값을 다시 저장합니다. Java의 list.add(...)처럼 쓰던 습관 때문에 append 결과를 버리지 않도록 주의하세요.

range는 반복할 때 인덱스와 값을 줍니다. 인덱스가 필요 없으면 for _, title := range titles처럼 _로 무시합니다.

## 3. slice를 다른 변수에 넣으면 항목은 공유될 수 있어요

```go
package main

import "fmt"

func main() {
    original := []int{1, 2}
    alias := original
    alias[0] = 9
    fmt.Println(original)

    copied := make([]int, len(original))
    copy(copied, original)
    copied[0] = 7
    fmt.Println(original)
    fmt.Println(copied)
}
```

```text
[9 2]
[9 2]
[7 2]
```

alias := original만으로 정수들이 담긴 배열을 새로 만들지는 않습니다. 두 slice가 같은 내부 배열을 보고 있어서 alias[0]을 바꾸면 original에서도 바뀐 값이 보입니다.

독립적인 정수 모음이 필요해 make로 공간을 만들고 copy로 항목을 복사했습니다. 항목이 포인터나 다른 slice라면 그 안의 대상까지 모두 복사하는 것은 아닙니다. 지금 예시는 int만 담아 복사 효과를 확인합니다.

append가 언제 새 내부 배열을 만드는지와 capacity는 나중에 읽어도 됩니다. 지금은 'slice를 대입했다고 항목까지 독립적으로 복사되지는 않는다'를 기억하세요.

## 4. map: 키로 값을 찾아요

```go
package main

import "fmt"

func main() {
    counts := map[string]int{}
    counts["go"] = 2

    value, exists := counts["go"]
    fmt.Println(value, exists)

    missing, exists := counts["java"]
    fmt.Println(missing, exists)
}
```

```text
2 true
0 false
```

map[string]int는 '키는 string, 값은 int인 map'입니다. counts["go"] = 2로 저장합니다.

조회 결과를 값 하나만 받으면 없는 키도 int의 기본값인 0으로 나옵니다. value, exists처럼 두 값을 받으면 실제로 키가 있었는지 구분할 수 있습니다. exists라는 이름은 직접 정한 변수명입니다. ok를 쓰는 코드도 자주 보게 됩니다.

처음 만들 때는 예시처럼 {}를 붙이거나 make(map[string]int)를 쓰세요. var counts map[string]int만 적으면 nil map이 되어 읽을 수는 있어도 항목을 쓸 때 panic이 납니다. map의 반복 순서도 보장되지 않습니다.

## 5. 저장소 코드에서 찾아보세요

[기본 예제](../examples/01-basics/main.go)의 다음 부분입니다.

```go
articles := []Article{article, {Title: "Context"}}
counts := map[bool]int{}
for _, item := range articles {
    counts[item.Read]++
}
```

articles는 Article을 담는 slice이고, counts는 true와 false를 키로 사용하는 map입니다. 읽은 항목 수와 안 읽은 항목 수를 따로 셉니다. 중첩된 {Title: "Context"}의 Article 타입은 바깥 slice 타입에서 알 수 있어 생략한 것입니다.

Article의 정확한 정의는 다음 장에서 배웁니다. 지금은 '두 항목을 반복하면서 bool 값별로 센다'고 읽으면 됩니다.

## 직접 확인해 보세요

1. titles에 문서를 하나 더 추가하면 len은 얼마가 되나요?
2. 없는 map 키를 조회한 값 0과 실제로 저장된 0은 어떻게 구분하나요?
3. alias[0]을 수정했을 때와 copied[0]을 수정했을 때 결과가 다른 이유를 설명해 보세요.

[이전: 실행과 기본 문법](01-basics.md) · [다음: 구조체·메서드·포인터](03-struct-methods.md)

더 읽기: [Go 공식 slice 설명](https://go.dev/blog/slices-intro).
