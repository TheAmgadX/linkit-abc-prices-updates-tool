package gui

import "embed"

// Assets holds the embedded frontend files served by the Wails asset server.
// The embed path is relative to this file (gui/embed.go), so "frontend" is correct.
//
//go:embed frontend
var Assets embed.FS
