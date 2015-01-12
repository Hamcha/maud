editorAdd = (elem, tag) ->
	txt = elem.parentElement.parentElement.text
	cursor = txt.selectionStart
	selectionLen = txt.selectionEnd - txt.selectionStart
	txt.value = 	(if txt.selectionStart > 0 then txt.value[0..txt.selectionStart-1] else "") +
			"[#{tag}]" + txt.value[txt.selectionStart..txt.selectionEnd-1] + "[/#{tag}]" +
			txt.value[txt.selectionEnd..]
	txt.selectionStart = cursor + tag.length + 2
	txt.selectionEnd = txt.selectionStart + selectionLen
	txt.focus()

window.editorAdd = editorAdd
