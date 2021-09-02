package sprinter

import (
	"fmt"
	"net/http"
	"testing"
)

func TestCrawler(t *testing.T) {
	testFixtures := []struct{
		content string
		tree *Node
	}{
		{
			content: "",
			tree: NewNode("http://127.0.0.1:8080/0"),
		},
		{
			content: `<a><img src="http://example.com/image-file.png" /></a>`,
			tree: NewNode("http://127.0.0.1:8080/1"),
		},
		{
			content: `<a href="http://127.0.0.1:8080/0">anchored text</a>`,
			tree: &Node{
				pageLinks: []string{"http://127.0.0.1:8080/0"},
				link:      "http://127.0.0.1:8080/2",
				child:     []*Node{
					{
						pageLinks: []string{},
						link:      "http://127.0.0.1:8080/0",
						child:     nil,
					},
				},
				},
		},
		{
			content: `<a HREF="http://127.0.0.1:8080/0">anchored text</a>`,
			tree: &Node{
				pageLinks: []string{"http://127.0.0.1:8080/0"},
				link:      "http://127.0.0.1:8080/3",
				child:     []*Node{
					{
						pageLinks: []string{},
						link:      "http://127.0.0.1:8080/0",
						child:     nil,
					},
				},
			},
		},
	}

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()

	for idx, tc := range testFixtures {
		handler := func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%s",tc.content)
		}

		http.HandleFunc(fmt.Sprintf("/%d",idx), handler)

		s, err := NewSprinter(fmt.Sprintf("http://127.0.0.1:8080/%d",idx))
		if err != nil {
			t.Fatal(err)
		}

		res := s.Crawl()
		if res.Link() != tc.tree.Link() {
			t.Fatalf("tree does not match : %s != %s", res.Link(), tc.tree.Link())
		}

		if res.String() != tc.tree.String() {
			t.Fatalf("tree does not match : %s != %s", res.String(), tc.tree.String())
		}
	}
}