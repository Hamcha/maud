editorAdd = (elem, tag) ->
	txt = elem.parentElement.parentElement.text
	cursor = txt.selectionStart
	selectionLen = txt.selectionEnd - txt.selectionStart
	txt.value = (if txt.selectionStart > 0 then txt.value[0..txt.selectionStart-1] else "") +
		"[#{tag}]" + txt.value[txt.selectionStart..txt.selectionEnd-1] + "[/#{tag}]" +
		txt.value[txt.selectionEnd..]
	txt.selectionStart = cursor + tag.length + 2
	txt.selectionEnd = txt.selectionStart + selectionLen
	txt.focus()

window.editorAdd = editorAdd

quoteText = (elem) ->
	txt = elem.parentElement.parentElement.text
	if txt.selectionStart == txt.selectionEnd
		txt.value = txt.value[0..txt.selectionStart] + "> " + txt.value[txt.selectionStart+1..]
		txt.focus()
		return
	txt.value = (if txt.selectionStart > 0 then txt.value[0..txt.selectionStart-1] else "") + "> " +
		txt.value[txt.selectionStart..txt.selectionEnd].trim().replace(///\n///g, "\n> ") + "\n" +
		txt.value[txt.selectionEnd+1..]
	txt.selectionStart = txt.selectionEnd = txt.value.length + 1
	txt.focus()

window.quoteText = quoteText
