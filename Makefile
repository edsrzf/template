# Copyright 2011 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=github.com/edsrzf/template
GOFILES=\
	filter.go\
	iftag.go\
	lexer.go\
	parser.go\
	tag.go\
	template.go\
	value.go\

include $(GOROOT)/src/Make.pkg
