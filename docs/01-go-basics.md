# 01. Go 기본 문법

## 패키지, 모듈, 실행 파일

모듈은 `go.mod`로 버전과 의존성을 관리하는 단위다. 패키지는 같은 디렉터리의 Go 파일들을 묶는 단위다. 서로 import할 수 있지만 순환 import는 허용되지 않는다. 실행 파일은 `package main`의 `func main()`에서 시작한다. 이 저장소는 모듈 하나에 실행 파일 세 개를 둔다.

`cmd/reading`이 실행 시작점이고 `internal/reading`이 목록 기능이다. `internal`에는 상위 경로를 공유하는 코드만 import할 수 있다는 언어 도구의 제한이 있다. `cmd`라는 이름 자체는 관례다. `pkg`, `controller`, `repository` 폴더가 있어야 Go 프로젝트가 되는 것은 아니다.

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

`reading.Store`는 서비스가 필요한 Create, Get, SetRead만 정의한다. MySQL 구현체가 이 메서드를 제공하므로 사용할 수 있다. 테스트도 같은 계약을 만족하는 작은 대역을 전달한다. 구현체마다 관성적으로 interface를 만드는 것은 필요하지 않다.

interface 값에는 동적 타입과 값이 함께 들어간다. `(*SomeError)(nil)`을 error로 반환하면 interface 자체는 nil이 아닐 수 있다. 성공할 때는 명확히 `return nil`을 쓴다.

## error와 defer

Go의 일반적인 실패 전달은 `(결과, error)`다. 예외가 자동으로 상위 호출자를 건너뛰는 방식이 아니므로 호출자가 실패 처리를 드러낸다.

```go
if err != nil {
    return fmt.Errorf("insert reading item: %w", err)
}
```

`%w`는 원래 오류를 감싼다. `errors.Is`는 특정 오류인지, `errors.As`는 특정 오류 타입인지 검사한다. 문자열 비교로 MySQL 중복 오류를 판단하지 않고 `MySQLError.Number`를 보는 이유다. `panic`은 예상 가능한 잘못된 사용자 입력 처리에 쓰지 않는다.

`defer`는 현재 함수가 반환될 때 실행한다. 파일 닫기, DB Rows 닫기, 트랜잭션 Rollback처럼 빠지는 경로가 생기면 안 되는 정리에 사용한다. 여러 defer는 역순으로 실행된다. 반복문에서 함수가 끝날 때까지 수많은 자원을 붙잡지 않도록 범위를 작게 만든다.

## Spring과 비교하며 볼 부분

| 익숙한 개념 | 이 예제의 Go 방식 |
| --- | --- |
| 컨테이너가 의존성 주입 | main에서 실제 객체를 만들고 생성 함수에 전달 |
| annotation 트랜잭션 | BeginTx, tx 메서드, Commit, Rollback을 명시 |
| 예외 전달 | error 반환, 감싸기와 분류 |
| 클래스와 상속 | struct, 메서드, 작은 interface와 조합 |
| entity 자동 추적 | SQL과 Scan으로 읽기/쓰기 명시 |

어느 방식이 무조건 낫다는 비교가 아니다. Go 코드에서 의존성, 실패와 자원 수명이 어느 줄에 드러나는지 찾는다.

## 확인할 것

- slice를 인자로 넘긴 뒤 항목을 바꾸면 왜 호출자도 바뀌는가?
- `Store`가 `mysqlstore`가 아니라 `reading`에 있는 이유는?
- `defer tx.Rollback()`을 두고 성공 경로에서 Commit하는 이유는?

읽기: [Tour of Go](https://go.dev/tour/), [공식 모듈 구조 안내](https://go.dev/doc/modules/layout), [Code Review Comments](https://go.dev/wiki/CodeReviewComments).
