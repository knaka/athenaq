// Code generated by gen-webui-const. DO NOT EDIT.

package main

import (
	"fmt"

	"github.com/webui-dev/go-webui/v2"
)

func StrToBrowser(s string) (browser webui.Browser, err error) {
	switch s {
	case "NoBrowser":
		browser = webui.NoBrowser
	case "AnyBrowser":
		browser = webui.AnyBrowser
	case "Chrome":
		browser = webui.Chrome
	case "Firefox":
		browser = webui.Firefox
	case "Edge":
		browser = webui.Edge
	case "Safari":
		browser = webui.Safari
	case "Chromium":
		browser = webui.Chromium
	case "Opera":
		browser = webui.Opera
	case "Brave":
		browser = webui.Brave
	case "Vivaldi":
		browser = webui.Vivaldi
	case "Epic":
		browser = webui.Epic
	case "Yandex":
		browser = webui.Yandex
	case "ChromiumBased":
		browser = webui.ChromiumBased
	default:
		err = fmt.Errorf("invalid browser: %s", s)
	}
	return
}
