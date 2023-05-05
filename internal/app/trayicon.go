// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package app

type trayIcon struct{}

func (icon trayIcon) Name() string {
	return "TrayIcon"
}

func (icon trayIcon) Content() []byte {
	return defaultIconData
}

type disabledIcon struct{}

func (icon disabledIcon) Name() string {
	return "TrayIcon"
}

func (icon disabledIcon) Content() []byte {
	return disabledIconData
}

type notifyingIcon struct{}

func (icon notifyingIcon) Name() string {
	return "TrayIcon"
}

func (icon notifyingIcon) Content() []byte {
	return disabledIconData
}
