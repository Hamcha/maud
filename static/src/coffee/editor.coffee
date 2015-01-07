editorAdd = (elem, tag) ->
	txt = elem.parentElement.parentElement.text
	txt.value += "[#{tag}][/#{tag}]"
	txt.selectionStart = txt.selectionEnd = txt.value.length - 3 - tag.length
	txt.focus()

window.editorAdd = editorAdd