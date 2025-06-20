package websiteinfos

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

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
func Get(str_url string) (*WebsiteInfos, error) {
	// Force HTTPS
	u, err := url.Parse(str_url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	u.Scheme = "https"
	str_url = u.String()

	// verify URL
	err = verifyURLIsSafe(str_url)
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
		if icon != nil {
			logrus.Infof("Found potential icon: %v", icon)
			// check if valid
			if !isValidPWAIcon(*icon) {
				icon = nil
			}
		}

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

	// Print results
	if title == "" {
		logrus.Warnf("Failed to find title in HTML")
	}
	var iconURL string
	if icon == nil {
		logrus.Warnf("Failed to find icon in HTML")
	} else {
		iconURL = icon.link
	}

	// Return
	return &WebsiteInfos{
		Title:   title,
		IconURL: iconURL,
	}, nil

}
