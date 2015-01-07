editorAdd = (elem, tag) ->
    txt = elem.parentElement.parentElement.text
    cursor = txt.selectionStart
    txt.value = txt.value[0..txt.selectionStart-1] + "[#{tag}][/#{tag}]" + txt.value[txt.selectionStart..]
    txt.selectionStart = txt.selectionEnd = cursor + tag.length + 2
    txt.focus()

window.editorAdd = editorAdd
