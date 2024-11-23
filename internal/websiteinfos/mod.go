package websiteinfos

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

type WebsiteInfos struct {
	Title   string `json:"title"`
	IconURL string `json:"icon_url"`
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

type icon struct {
	height int
	width  int
	link   string
}

func (i *icon) AnySize() bool {
	return i.height == 0 && i.width == 0
}

func sizeScore(height, width int) int {
	return height * width
}

func parseSize(size string) (int, int) {
	split := strings.Split(size, "x")
	if len(split) != 2 {
		logrus.Warnf("Found invalid size property: %v", size)
		return 0, 0
	}

	height, err := strconv.Atoi(split[0])
	if err != nil {
		logrus.Warnf("Failed to parse height: %v", err)
		return 0, 0
	}

	width, err := strconv.Atoi(split[1])
	if err != nil {
		logrus.Warnf("Failed to parse width: %v", err)
		return 0, 0
	}

	return height, width
}

// Construct a icon struct that represents the best size this icon can have
func constructIconStruct(base_link, rel_link string, sizes string) *icon {
	link, err := url.JoinPath(base_link, rel_link)
	if err != nil {
		logrus.Warnf("Failed to join URL path: %v", err)
		return nil
	}

	currentIcon := icon{}

	for _, size := range strings.Split(sizes, " ") {
		if size == "any" {
			return &icon{
				link: link,
			}
		}

		height, width := parseSize(size)
		if height == 0 && width == 0 {
			continue
		}

		if sizeScore(height, width) > sizeScore(currentIcon.height, currentIcon.width) {
			currentIcon = icon{
				height: height,
				width:  width,
				link:   link,
			}
		}
	}

	if currentIcon.height == 0 && currentIcon.width == 0 {
		return nil
	} else {
		return &currentIcon
	}
}

func tryFindIcon(str_url string, n *html.Node) *icon {
	// Search for basic icons
	if n.Type == html.ElementNode && n.Data == "link" {
		for _, attr := range n.Attr {
			// Look for rel
			if attr.Key != "rel" {
				continue
			}

			// check if icon or apple-touch-icon
			values := strings.Split(attr.Val, " ")
			if !slices.Contains(values, "icon") && !slices.Contains(values, "apple-touch-icon") {
				continue
			}

			href := getAttr(n, "href")
			sizes := getAttr(n, "sizes")
			if href == nil || sizes == nil {
				logrus.Warningf("Found icon without href or sizes: %v", nodeToString(n))
				continue
			}

			// Extract icon
			icon := constructIconStruct(str_url, href.Val, sizes.Val)
			if icon != nil {
				return icon
			}
		}
	}

	return nil
}

func Get(str_url string) (*WebsiteInfos, error) {
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
	var title string
	var icon *icon
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
			if title != "" && icon != nil {
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
	if icon == nil {
		return nil, fmt.Errorf("failed to find icon in HTML")
	}

	// Return
	return &WebsiteInfos{
		Title:   title,
		IconURL: icon.link,
	}, nil

}
