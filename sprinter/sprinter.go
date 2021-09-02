package sprinter

import (
	"errors"
	"fmt"
	"github.com/temoto/robotstxt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var errInvalidURL = errors.New("url is invalid")

// Sprinter is a web crawler that will crawl a root domain and collect all child links from the root.
// It is limited to a subdomain and will not follow external links.
type Sprinter struct {
	group  *robotstxt.Group
	root    string
	mu      sync.Mutex // Locks read/write to visited map
	visited map[string]bool
	jobs    sync.WaitGroup
	errChan chan error
}

// NewSprinter returns a new sprinter.
func NewSprinter(root string) (s *Sprinter, err error) {
	if !validURL(root) {
		return nil, fmt.Errorf("%s: got '%s'", errInvalidURL, root)
	}

	root, err = cleanURL(root, root)
	if err != nil {
		return nil, err
	}

	return &Sprinter{
		root:    root,
		errChan: make(chan error),
	}, nil
}

// Root returns the root domain that is will be craweled.
func (s *Sprinter) Root() string {
	return s.root
}

// Crawl will start the sprinter and return the list of links.
func (s *Sprinter) Crawl() *Node {
	// visited is a map of urls visited
	s.visited = make(map[string]bool)

	go func() {
		for err := range s.errChan {
			log.Printf("Recieved error : %s", err)
		}
	}()

	// First, check the robots.txt.
	resp, err := http.Get(fmt.Sprintf("%s/robots.txt", s.Root()))
	if err != nil {
		s.errChan <- fmt.Errorf("unable to get robots.txt : %s\nshutting down...\n", err)
		return nil
	}

	robots, err := robotstxt.FromResponse(resp)
	if err != nil {
		s.errChan <- fmt.Errorf("unable to parse robots.txt : %s", err)
		return nil
	}

	s.group = robots.FindGroup("SprinterBot")

	root := s.visit(NewNode(s.root))
	s.jobs.Wait()
	return root
}

func (s *Sprinter) setVisited(u string) {
	s.mu.Lock()
	s.visited[u] = true
	s.mu.Unlock()
}
func (s *Sprinter) getVisited(u string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, visited := s.visited[u]
	return visited
}

// Visit does the following.
// 1. Request the HTML for the given resource in node.
func (s *Sprinter) visit(root *Node) (nodes *Node) {
	if root == nil {
		return
	}

	isVisited := s.getVisited(root.Link())
	if isVisited {
		return nil
	}

	if !s.group.Test(root.Link()) {
		s.errChan <- fmt.Errorf("visiting %s violates robots.txt", root.Link())
	//	 add return
	}

	s.setVisited(root.Link())

	// TODO Use context..
	resp, err := http.Get(root.Link())
	if err != nil {
		s.errChan <- fmt.Errorf("unable to get %s : %s", root.Link(), err)
		return nodes
	}

	if resp.StatusCode != http.StatusOK {
		s.errChan <- fmt.Errorf("received %s for %s", resp.Status, root.Link())
		return nodes
	}

	cursor := root

	links := parseLinks(resp.Body)

	links = filterLinks(links)

	children := make([]string, 0)

	for _, link := range links {
		// Clean the URL.
		link, err = cleanURL(link, root.Link())
		if err != nil {
			s.errChan <- fmt.Errorf("issue cleaning %s : %s", link, err)
			continue
		}

		// If the link is cyclical, do not include.
		if link == cursor.Link() {
			continue
		}

		cursor.pageLinks = append(cursor.pageLinks, link)

		// Check if we should continue.
		if !shouldCrawl(link, s.root) {
			continue
		}

		// If we haven't visited the link, add it.
		s.mu.Lock()
		_, visited := s.visited[link]
		if !visited {
			children = append(children, link)
		}
		s.mu.Unlock()
	}

	for _, child := range children {
		s.jobs.Add(1)
		go func(child string) {
			node := s.visit(NewNode(child))
			if node != nil {
				cursor.child = append(cursor.child, node)
			}
			s.jobs.Done()
		}(child)
	}

	return root
}

func filterLinks(links []string) []string {
	// Remove fragment identifiers from the links.
	for idx, link := range links {
		url := strings.Split(link, "#")
		links[idx] = url[0]
	}

	// Only iterate through *unique* links.
	{
		linkSet := make(map[string]bool)
		for _, link := range links {
			linkSet[link] = true
		}

		links = make([]string, len(linkSet))

		idx := 0
		for link := range linkSet {
			links[idx] = link
			idx++
		}
	}

	return links
}

// validURL sanity checks the provided URL is valid.
func validURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	_, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return true
}

// parseLinks returns a slice of URLS on a given page.
func parseLinks(r io.Reader) (links []string) {
	// We use golang.org/x/net's html library to parse an HTML resource.
	d := html.NewTokenizer(r)
	for tt := d.Next(); tt != html.ErrorToken; tt = d.Next() {
		switch tt {
		case html.StartTagToken, html.EndTagToken:
			token := d.Token()
			if "a" == strings.ToLower(token.Data) {
				for _, attr := range token.Attr {
					if strings.ToLower(attr.Key) == "href" {
						links = append(links, strings.ToLower(attr.Val))
					}
				}
			}
		}
	}
	return links
}

// TODO, find sitemap
// TODO, read robots.txt

func shouldCrawl(next string, root string) bool {
	// TODO add subdomain, tld, lowercase...
	uri, err := url.Parse(next)
	if err != nil {
		return false
	}

	domain, err := url.Parse(root)
	if err != nil {
		return false
	}

	// Must be same host.
	if uri.Hostname() != domain.Hostname() {
		return false
	}

	return true
}

func cleanURL(u, base string) (res string, err error) {
	resource, err := url.Parse(u)
	if err != nil {
		return res, err
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return res, err
	}
	uri := baseURL.ResolveReference(resource)

	return uri.String(), err
}
