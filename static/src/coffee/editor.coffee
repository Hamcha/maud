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

quoteText = (elem) ->
	txt = elem.parentElement.parentElement.text
	if txt.selectionStart == txt.selectionEnd
		txt.value = txt.value[0..txt.selectionStart] + "> " + txt.value[txt.selectionStart+1..]
		txt.focus()
		return
	txt.value = (if txt.selectionStart > 0 then txt.value[0..txt.selectionStart-1] else "") + "> " +
		txt.value[txt.selectionStart..txt.selectionEnd].trim().replace(/\n/g, "\n> ") + "\n" +
		txt.value[txt.selectionEnd+1..]
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
    <a onclick="Editor.quoteText(this)">&gt;</a>
"""

window.Editor =
	add: editorAdd
	quoteText: quoteText
