package main

import (
	"errors"
	"fmt"
	"strings"
)

var errEmptyTitle = errors.New("title is empty")

type Article struct {
	Title string
	Read  bool
}

func NewArticle(title string) (Article, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Article{}, fmt.Errorf("new article: %w", errEmptyTitle)
	}
	return Article{Title: title}, nil
}

func (a *Article) MarkRead()     { a.Read = true }
func (a Article) String() string { return fmt.Sprintf("%s (read=%t)", a.Title, a.Read) }

// Article satisfies fmt.Stringer without an implements declaration.
func printLabel(value fmt.Stringer) { fmt.Println(value.String()) }

func main() {
	article, err := NewArticle(" A Tour of Go ")
	if err != nil {
		panic(err)
	}
	article.MarkRead()
	printLabel(article)
	articles := []Article{article, {Title: "Context"}}
	counts := map[bool]int{}
	for _, item := range articles {
		counts[item.Read]++
	}
	fmt.Printf("read=%d unread=%d\n", counts[true], counts[false])
	_, err = NewArticle(" ")
	fmt.Println("empty title:", errors.Is(err, errEmptyTitle))
}
