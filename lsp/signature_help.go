// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lsp

import (
	"context"

	"golang.org/x/tools/internal/lsp/protocol"
	"golang.org/x/tools/internal/lsp/source"
	"golang.org/x/tools/internal/telemetry/log"
	"golang.org/x/tools/internal/telemetry/tag"
)

func (s *Server) signatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	uri := params.TextDocument.URI.SpanURI()
	view, err := s.session.ViewOf(uri)
	if err != nil {
		return nil, err
	}
	snapshot := view.Snapshot()
	fh, err := snapshot.GetFile(uri)
	if err != nil {
		return nil, err
	}
	if fh.Identity().Kind != source.Go {
		return nil, nil
	}
	info, activeParameter, err := source.SignatureHelp(ctx, snapshot, fh, params.Position)
	if err != nil {
		log.Print(ctx, "no signature help", tag.Of("At", params.Position), tag.Of("Failure", err))
		return nil, nil
	}
	return &protocol.SignatureHelp{
		Signatures:      []protocol.SignatureInformation{*info},
		ActiveParameter: float64(activeParameter),
	}, nil
}
