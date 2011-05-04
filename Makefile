# Copyright 2011 Evan Shaw. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=github.com/edsrzf/template
GOFILES=\
	expr.go\
	filter.go\
	lexer.go\
	parser.go\
	tag.go\
	template.go\
	value.go\

include $(GOROOT)/src/Make.pkg
