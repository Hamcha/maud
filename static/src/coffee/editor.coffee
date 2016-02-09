editorAdd = (elem, tag) ->
	# get the textarea and the selection
	txt = elem.parentElement.parentElement.text
	cursor = txt.selectionStart
	selectionLen = txt.selectionEnd - txt.selectionStart
	text = txt.value
	[start, end] = [txt.selectionStart, txt.selectionEnd]
	txt.value = if start is 0 then "" else text[..start-1]
	txt.value += "[#{tag}]#{if end > 0 then text[start..end-1] else ""}[/#{tag}]#{text[end..]}"
	txt.selectionStart = cursor + tag.length + 2
	txt.selectionEnd = txt.selectionStart + selectionLen
	txt.focus()

quoteText = (elem) ->
	txt = elem.parentElement.parentElement.text
	unless window.getSelection()?.toString()?.length > 0
		txt.value = txt.value[0..txt.selectionStart] + "> " + txt.value[txt.selectionStart+1..]
		txt.focus()
		return
	separator = if txt.selectionStart > 0 and txt.value[txt.selectionStart - 1] isnt "\n" then "\n" else ""
	txt.value = txt.value[0...txt.selectionStart] + separator + "> #{window.getSelection()}\n" + txt.value[txt.selectionEnd+1..]
	txt.selectionStart = txt.selectionEnd = txt.value.length + 1
	txt.focus()

# Setup editor buttons
document.getElementById('editorButtons')?.innerHTML = """
    <a onclick="Editor.add(this, 'b')"><b>B</b></a>
    <a onclick="Editor.add(this, 'i')"><i>i</i></a>
    <a onclick="Editor.add(this, 'u')"><u>u</u></a>
    <a onclick="Editor.add(this, 's')"><s>strike</s></a>
    <a onclick="Editor.add(this, 'img')">img</a>
    <a onclick="Editor.add(this, 'url')"><span style="border-bottom: 1px dotted #fff">url</span></a>
    <a onclick="Editor.add(this, 'spoiler')">spoiler</a>
    <a onclick="Editor.add(this, 'youtube')">youtube</a>
    <a onclick="Editor.add(this, 'html')">html</a>
    <a onclick="Editor.add(this, 'video')">video</a>
    <a onclick="Editor.add(this, 'pre')">pre</a>
    <a onmousedown="Editor.quoteText(this)">&gt;</a>
"""

window.Editor =
	add: editorAdd
	quoteText: quoteText
