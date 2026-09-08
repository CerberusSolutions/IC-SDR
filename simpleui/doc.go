// Package simpleui implements a scalable retained-mode user interface on top
// of raylib-go. Applications draw in a stable logical coordinate system while
// Canvas and Viewport map rendering and input to a resizable host window.
//
// The package is intentionally independent from package main so it can be
// reused by other applications in this module and later extracted as its own
// Go module if needed.
package simpleui
