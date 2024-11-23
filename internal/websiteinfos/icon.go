package websiteinfos

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

type iconType int64

const (
	None        iconType = 0
	Normal      iconType = 1
	AnySize     iconType = 2
	UnknownSize iconType = 3
)

type icon struct {
	height int
	width  int
	link   string
	mode   iconType
}

func (i *icon) String() string {
	return fmt.Sprintf("Icon{height: %v, width: %v, link: %v, mode: %v}", i.height, i.width, i.link, i.mode)
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
func guessIconSize(link string, sizes string) icon {
	// start with unknown size by default
	currentIcon := icon{
		link:   link,
		height: 0,
		width:  0,
		mode:   UnknownSize,
	}

	for _, size := range strings.Split(sizes, " ") {
		// Check size attribute
		// if any size, don't look any further, we got the best possible
		if size == "any" {
			return icon{
				link: link,
				mode: AnySize,
			}
		}

		height, width := parseSize(size)
		if height == 0 && width == 0 { // unknown size
			continue
		}

		if sizeScore(height, width) > sizeScore(currentIcon.height, currentIcon.width) {
			currentIcon = icon{
				mode:   Normal,
				height: height,
				width:  width,
				link:   link,
			}
		}
	}

	// if size is still unknown, check if it's a svg
	if currentIcon.height == 0 && currentIcon.width == 0 {
		if strings.HasSuffix(link, ".svg") {
			return icon{
				link: link,
				mode: AnySize,
			}
		}
	}
	return currentIcon
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

		// Compute absolute link
		link, err := url.JoinPath(str_url, href.Val)
		if err != nil {
			logrus.Warnf("Failed to join URL path: %v", err)
			return nil
		}

		// Extract icon size
		icon := guessIconSize(link, sizes.Val)
		return &icon
	}

	return nil
}

func isValidPWAIcon(i icon) bool {
	if i.mode == AnySize {
		return true
	}

	if i.mode == Normal {
		return i.height >= 144 && i.width >= 144
	}

	return false
}
