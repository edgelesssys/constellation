/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logger

import (
	"context"
	"log/slog"
)

// LevelHandler copied from the official LevelHandler example in the slog package documentation.

// levelHandler wraps a Handler with an Enabled method
// that returns false for levels below a minimum.
type levelHandler struct {
	level   slog.Leveler
	handler slog.Handler
}

// newLevelHandler returns a LevelHandler with the given level.
// All methods except Enabled delegate to h.
func newLevelHandler(level slog.Leveler, h slog.Handler) *levelHandler {
	// Optimization: avoid chains of LevelHandlers.
	if lh, ok := h.(*levelHandler); ok {
		h = lh.Handler()
	}
	return &levelHandler{level, h}
}

// Enabled implements Handler.Enabled by reporting whether
// level is at least as large as h's level.
func (h *levelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements Handler.Handle.
func (h *levelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements Handler.WithAttrs.
func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newLevelHandler(h.level, h.handler.WithAttrs(attrs))
}

// WithGroup implements Handler.WithGroup.
func (h *levelHandler) WithGroup(name string) slog.Handler {
	return newLevelHandler(h.level, h.handler.WithGroup(name))
}

// Handler returns the Handler wrapped by h.
func (h *levelHandler) Handler() slog.Handler {
	return h.handler
}
