package intmain

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

type WebsiteInfos struct {
	Title string `json:"title"`
	IconURL  string `json:"icon_url"`
}

// Verify we aren't being tricked by a malicious actor
// Deny private IPs, and everything not HTTPS
func verifyURLIsSafe(url_str string) error {
	if len(url_str) > 256 {
		return fmt.Errorf("URL is too long")
	}

	u, err := url.Parse(url_str)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("URL scheme is not HTTPS")
	}

	// Lookup if IP is private
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return fmt.Errorf("failed to lookup IP address: %w", err)
	}
	for _, ip := range ips {
		if ip.IsPrivate() {
			return fmt.Errorf("URL resolves to a private IP address")
		}
		if ip.IsLoopback() {
			return fmt.Errorf("URL resolves to a loopback IP address")
		}
		if ip.IsMulticast() {
			return fmt.Errorf("URL resolves to a multicast IP address")
		}
		if ip.IsUnspecified() {
			return fmt.Errorf("URL resolves to an unspecified IP address")
		}
	}

	return nil

}

func tryFindIcon(str_url string, n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "link" {
		for _, attr := range n.Attr {
			if attr.Key == "rel" && slices.Contains(strings.Split(attr.Val, " "), "icon") {
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						icon, err := url.JoinPath(str_url, attr.Val)
						if err != nil {
							fmt.Printf("Failed to join URL path: %v\n", err)
						} else {
							return icon
						}
					}
				}
			}
		}
	}

	return ""
}

func getWebsiteInfos(str_url string) (*WebsiteInfos, error) {
	// verify URL
	err := verifyURLIsSafe(str_url)
	if err != nil {
		return nil, fmt.Errorf("URL is not safe: %w", err)
	}

	// Make request
	resp, err := http.Get(str_url)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request to website: %w", err)
	}
	defer resp.Body.Close()

	// Parse
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Search HTML for title and icon
	var title, icon string
	var f func(*html.Node)
	f = func(n *html.Node) {

		// search title
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
		}

		// search icon
		icon = tryFindIcon(str_url, n)

		// Process childs
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// stop searching if we found both
			if title != "" && icon != "" {
				return
			}

			f(c)
		}
	}
	f(doc)

	// Verify
	if title == "" {
		return nil, fmt.Errorf("failed to find title in HTML")
	}
	if icon == "" {
		return nil, fmt.Errorf("failed to find icon in HTML")
	}

	// Return
	return &WebsiteInfos{
		Title: title,
		IconURL:  icon,
	}, nil

}
