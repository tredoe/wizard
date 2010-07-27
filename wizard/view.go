// Copyright 2010, The 'gowizard' Authors.  All rights reserved.
// Use of this source code is governed by the 2-clause BSD License
// that can be found in the LICENSE file.

package wizard

import (
	"fmt"
	"os"

	"github.com/hoisie/mustache.go"
	"github.com/hoisie/web.go"
)


type view struct {
	//url      string
	template string
	rendered string            // Rendered template
	tag      map[string]string // Values to send to the template
}

func NewView(template string, tag map[string]string) *view {
	_view := new(view)
	_view.template = fmt.Sprint(_DIR_TEMPLATE, template)
	_view.tag = tag

	return _view
}

/* Renders the template using Mustache. */
func (self *view) render(ctx *web.Context) {
	if self.rendered == "" {
		// Renders the template
		var err os.Error

		self.rendered, err = mustache.RenderFile(self.template, self.tag)
		if err != nil {
			ctx.Abort(500, fmt.Sprint("Unable to parse template file:", err.String()))
		}
	}
	ctx.WriteString(self.rendered)
}

