// Lute - 一款结构化的 Markdown 引擎，支持 Go 和 JavaScript
// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package parse

import (
	"bytes"

	"github.com/88250/lute/ast"
	"github.com/88250/lute/lex"
	"github.com/88250/lute/util"
)

func (t *Tree) parseFileAnnotationRef(ctx *InlineContext) *ast.Node {
	if !t.Context.ParseOption.FileAnnotationRef {
		return nil
	}

	tokens := ctx.tokens[ctx.pos:]
	if 48 > len(tokens) || lex.ItemLess != tokens[0] || lex.ItemLess != tokens[1] {
		return nil
	}

	var id, text []byte
	ctx.pos += 2
	var ok, matched bool
	var passed, remains []byte
	for { // 这里使用 for 是为了简化逻辑，不是为了循环
		if ok, passed, remains = lex.Spnl(ctx.tokens[ctx.pos:]); !ok {
			break
		}
		ctx.pos += len(passed)
		if passed, remains, id = t.Context.parseFileAnnotationRefID(remains); 1 > len(passed) {
			break
		}
		ctx.pos += len(passed)
		matched = lex.ItemGreater == passed[len(passed)-1] && lex.ItemGreater == passed[len(passed)-2]
		if matched {
			break
		}
		if 1 > len(remains) || !lex.IsWhitespace(remains[0]) {
			break
		}
		// 跟空格的话后续尝试 title 解析
		if ok, passed, remains = lex.Spnl(remains); !ok {
			break
		}
		ctx.pos += len(passed) + 1
		matched = 2 <= len(remains) && lex.ItemGreater == remains[0] && lex.ItemGreater == remains[1]
		if matched {
			ctx.pos++
			break
		}
		var validTitle bool
		if validTitle, passed, remains, text = t.Context.parseLinkTitle(remains); !validTitle {
			break
		}
		ctx.pos += len(passed)
		ok, passed, remains = lex.Spnl(remains)
		ctx.pos += len(passed)
		matched = ok && 1 < len(remains)
		if matched {
			matched = lex.ItemGreater == remains[0] && lex.ItemGreater == remains[1]
			ctx.pos += 2
		}
		break
	}
	if !matched {
		return nil
	}

	ret := &ast.Node{Type: ast.NodeFileAnnotationRef}
	ret.AppendChild(&ast.Node{Type: ast.NodeLess})
	ret.AppendChild(&ast.Node{Type: ast.NodeLess})
	ret.AppendChild(&ast.Node{Type: ast.NodeFileAnnotationRefID, Tokens: id})
	if 0 < len(text) {
		ret.AppendChild(&ast.Node{Type: ast.NodeFileAnnotationRefSpace})
		textNode := &ast.Node{Type: ast.NodeFileAnnotationRefText, Tokens: text}
		ret.AppendChild(textNode)
	}
	ret.AppendChild(&ast.Node{Type: ast.NodeGreater})
	ret.AppendChild(&ast.Node{Type: ast.NodeGreater})
	return ret
}

func (context *Context) parseFileAnnotationRefID(tokens []byte) (passed, remains, id []byte) {
	remains = tokens
	length := len(tokens)
	if 1 > length {
		return
	}

	var i int
	var token byte
	for ; i < length; i++ {
		token = tokens[i]
		if bytes.Contains(util.CaretTokens, []byte{token}) {
			continue
		}

		if bytes.HasPrefix(tokens[i:], []byte(" \"")) {
			break
		}

		if '>' == token {
			break
		}
	}
	remains = tokens[i:]
	id = tokens[:i]
	if 6 > len(remains) {
		return
	}
	passed = make([]byte, 0, 1024)
	passed = append(passed, id...)
	if bytes.HasPrefix(remains, util.CaretTokens) {
		passed = append(passed, util.CaretTokens...)
		remains = remains[len(util.CaretTokens):]
	}
	closed := lex.ItemGreater == remains[0] && lex.ItemGreater == remains[1]
	if closed {
		passed = append(passed, []byte(">>")...)
		return
	}

	if !lex.IsWhitespace(remains[0]) {
		passed = nil
		return
	}
	return
}
