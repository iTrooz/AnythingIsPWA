package websiteinfos

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

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
	// Search for icons from <img> tags
	if n.Type == html.ElementNode && n.Data == "link" {
		rel := getAttr(n, "rel")
		if rel == nil {
			return nil
		}

		// check if icon or apple-touch-icon
		values := strings.Split(rel.Val, " ")
		if !slices.Contains(values, "icon") &&
			!slices.Contains(values, "apple-touch-icon") &&
			!slices.Contains(values, "mask-icon") {
			return nil
		}

		href := getAttr(n, "href")
		sizes := getAttr(n, "sizes")
		if href == nil || sizes == nil {
			logrus.Warningf("Found icon without href or sizes: %v", nodeToString(n))
			return nil
		}

		// Extract icon
		icon := constructIconStruct(str_url, href.Val, sizes.Val)
		if icon != nil {
			return icon
		}
	}

	return nil
}
