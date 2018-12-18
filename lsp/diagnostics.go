// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lsp

import (
	"context"
	"sort"

	"golang.org/x/tools/internal/lsp/protocol"
	"golang.org/x/tools/internal/lsp/source"
)

func (s *server) cacheAndDiagnose(ctx context.Context, uri protocol.DocumentURI, content string) {
	sourceURI := fromProtocolURI(uri)
	if err := s.setContent(ctx, sourceURI, []byte(content)); err != nil {
		return // handle error?
	}
	go func() {
		reports, err := source.Diagnostics(ctx, s.view, sourceURI)
		if err != nil {
			return // handle error?
		}
		for filename, diagnostics := range reports {
			uri := source.ToURI(filename)
			s.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
				URI:         protocol.DocumentURI(uri),
				Diagnostics: toProtocolDiagnostics(ctx, s.view, uri, diagnostics),
			})
		}
	}()
}

func (s *server) setContent(ctx context.Context, uri source.URI, content []byte) error {
	v, err := s.view.SetContent(ctx, uri, content)
	if err != nil {
		return err
	}

	s.viewMu.Lock()
	s.view = v
	s.viewMu.Unlock()

	return nil
}

func toProtocolDiagnostics(ctx context.Context, v source.View, uri source.URI, diagnostics []source.Diagnostic) []protocol.Diagnostic {
	reports := []protocol.Diagnostic{}
	for _, diag := range diagnostics {
		f, err := v.GetFile(ctx, uri)
		if err != nil {
			continue // handle error?
		}
		tok, err := f.GetToken()
		if err != nil {
			continue // handle error?
		}
		reports = append(reports, protocol.Diagnostic{
			Message:  diag.Message,
			Range:    toProtocolRange(tok, diag.Range),
			Severity: protocol.SeverityError, // all diagnostics have error severity for now
			Source:   "LSP",
		})
	}
	return reports
}

func sorted(d []protocol.Diagnostic) {
	sort.Slice(d, func(i int, j int) bool {
		if d[i].Range.Start.Line == d[j].Range.Start.Line {
			if d[i].Range.Start.Character == d[j].Range.Start.Character {
				return d[i].Message < d[j].Message
			}
			return d[i].Range.Start.Character < d[j].Range.Start.Character
		}
		return d[i].Range.Start.Line < d[j].Range.Start.Line
	})
}
