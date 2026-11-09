# 01. Go 기본 문법

## 패키지, 모듈, 실행 파일

모듈은 `go.mod`로 버전과 의존성을 관리하는 단위다. 패키지는 같은 디렉터리의 Go 파일들을 묶는 단위다. 서로 import할 수 있지만 순환 import는 허용되지 않는다. 실행 파일은 `package main`의 `func main()`에서 시작한다. go-basic은 독립 모듈 하나에 문법과 동시성 예제를 둔다. 외부 라이브러리는 사용하지 않는다.

각 examples 하위 디렉터리는 실행 가능한 package main이다. internal/workpool은 동시성 예제가 import하는 패키지다. internal에는 상위 경로를 공유하는 코드만 import할 수 있다는 언어 도구의 제한이 있다.

## 값과 제로 값

```go
var count int       // 0
var enabled bool    // false
var title string    // ""
name := "Go"        // 함수 안에서 선언과 추론
```

구조체도 필드의 제로 값으로 만들어진다. 생성자 문법은 없고 필요하면 `New` 함수를 만든다. 제로 값이 곧 유효한 업무 데이터라는 뜻은 아니다. 빈 제목을 거절하는 일은 서비스의 검증이다.

Go는 할당과 인자 전달에서 값을 복사한다. struct 전체도 값이다. pointer를 전달하면 주소를 공유한다. `Article`을 값으로 받은 메서드에서 필드를 바꾸면 복사본이 바뀐다. `*Article`을 받으면 원본을 바꿀 수 있다. 큰 객체라는 이유만으로 모든 값을 pointer로 바꾸기보다 변경 의미와 수명을 본다.

## array, slice, map

array는 길이가 타입의 일부인 고정 크기 값이다. slice는 배열 구간을 가리키는 정보이며 길이와 용량을 가진다.

```go
original := []int{1, 2, 3}
alias := original[:2]
alias[0] = 9             // original[0]도 9
copied := append([]int(nil), original...)
copied[0] = 7            // original과 별도 배열
```

slice를 복사해도 내부 배열까지 깊은 복사가 되지는 않는다. `append`가 기존 배열을 쓰는지 새 배열을 만드는지는 capacity에 따라 달라진다. 공유 메모리와 동시성을 다룰 때 중요하다. 함수가 반환한 slice를 외부에서 바꿔도 되는지도 정한다.

map은 키로 값을 찾는 자료구조다. `value, ok := m[key]`로 실제 존재 여부를 구분한다. 없는 키에서 나온 0과 실제 저장한 0은 다르다. nil map의 조회는 가능하지만 쓰기는 panic이다. `make(map[string]int)`로 만든다. map의 반복 순서는 보장되지 않고, 동시 읽기와 쓰기는 보호가 필요하다.

## struct, 메서드, interface

```go
type Article struct { Title string; Read bool }
func (a *Article) MarkRead() { a.Read = true }
```

메서드는 특정 타입에 붙은 함수다. 클래스 상속보다는 작은 타입과 조합으로 기능을 나눈다. interface는 필요한 메서드의 집합이며 명시적 `implements` 없이 만족한다.

기본 예제의 printLabel은 fmt.Stringer 인터페이스를 인자로 받는다. Article에 String() string 메서드가 있으므로 별도의 implements 선언 없이 전달할 수 있다. 필요한 메서드가 있는지로 인터페이스를 만족한다.

interface 값에는 동적 타입과 값이 함께 들어간다. `(*SomeError)(nil)`을 error로 반환하면 interface 자체는 nil이 아닐 수 있다. 성공할 때는 명확히 `return nil`을 쓴다.

## error와 defer

Go의 일반적인 실패 전달은 `(결과, error)`다. 예외가 자동으로 상위 호출자를 건너뛰는 방식이 아니므로 호출자가 실패 처리를 드러낸다.

```go
if err != nil {
    return fmt.Errorf("create article: %w", err)
}
```

`%w`는 원래 오류를 감싼다. `errors.Is`는 특정 오류인지, `errors.As`는 특정 오류 타입인지 검사한다. 기본 예제는 errors.Is로 빈 제목 오류인지 확인한다. `panic`은 예상 가능한 잘못된 사용자 입력 처리에 쓰지 않는다.

`defer`는 현재 함수가 반환될 때 실행한다. 파일 닫기나 타이머 정리처럼 빠지는 경로가 생기면 안 되는 정리에 사용한다. 여러 defer는 역순으로 실행된다. 반복문에서 함수가 끝날 때까지 수많은 자원을 붙잡지 않도록 범위를 작게 만든다.

## 확인할 것

- slice를 인자로 넘긴 뒤 항목을 바꾸면 왜 호출자도 바뀌는가?
- Article에 implements 선언이 없는데 fmt.Stringer로 전달할 수 있는 이유는?
- MarkRead의 수신자를 *Article에서 Article로 바꾸면 원본은 어떻게 되는가?

읽기: [Tour of Go](https://go.dev/tour/), [공식 모듈 구조 안내](https://go.dev/doc/modules/layout), [Code Review Comments](https://go.dev/wiki/CodeReviewComments).
