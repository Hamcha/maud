'use strict'

original = []

editorAdd = (elem, tag) ->
	return ->
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
	return ->
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
addEditorButtons = (container) ->
	elements = [
		{ tag:    "b",       text: "<b>B</b>" },
		{ tag:    "i",       text: "<i>i</i>" },
		{ tag:    "u",       text: "<u>u</u>" },
		{ tag:    "s",       text: "<s>strike</s>" },
		{ tag:    "img",     text: "img" },
		{ tag:    "url",     text: "<span style=\"border-bottom: 1px dotted #fff\">url</span>" },
		{ tag:    "spoiler", text: "spoiler" },
		{ tag:    "youtube", text: "youtube" },
		{ tag:    "html",    text: "html" },
		{ tag:    "video",   text: "video" },
		{ tag:    "pre",     text: "pre" },
		{ tag:    "video",   text: "video" },
		{ action: "quote",   text: "&gt;" },
	]
	for element in elements
		button = document.createElement 'a'
		if element.tag?
			button.addEventListener "click", editorAdd button, element.tag
		if element.action?
			switch element.action
				when "quote"
					button.addEventListener "click", quoteText(button)
				else throw new Exception "Unknown editor op: " + element.action
		button.innerHTML = element.text
		container.appendChild button

window.onload = ->
	if AC
		AC.toggleAutocomplete (document.querySelector '#reply-form .ac_input'), '/taglist'

cancelForm = (id) ->
	return ->
		pid = if id == 0 then "thread" else "p#{id}"
		post = document.getElementById pid
		post.innerHTML = original[id]
		post.querySelector('.postEditLink')?.onclick = -> editPost +id
		post.querySelector('.postDeleteLink')?.onclick = -> deletePost +id

# post preview
showPreview = (where) ->
	return ->
		form = document.getElementById where
		text = (document.getElementById where + "-text").value
		unless text
			return if form.firstChild.className == 'errmsg'
			errmsg = window.createElementEx "p", { className: "errmsg", id: "#{where}-preview" }
			errmsg.appendChild document.createTextNode "Please insert some content."
			form.insertBefore errmsg, form.firstChild
			return
		req = { text: text }
		# retreive content data from the server
		qwest.post('/postpreview', req)
			.then (resp) ->
				createPreview where, resp
			.catch (err) ->
				return if form.firstChild.className == 'errmsg'
				errmsg = document.createElement 'p'
				errmsg.className = "errmsg"
				errmsg.innerHTML = "Failed to retrieve content: #{err}"
				form.insertBefore errmsg, form.firstChild

createPreview = (where, content) ->
	# deselect selected post, if any
	o.className = o.className.replace "post-selected", "" for o in document.querySelectorAll ".post-selected"
	# if preview post already exists, just update it
	prevpost = document.getElementById "#{where}-preview"
	unless prevpost?
		prevpost = window.createElementEx "article", { id: "#{where}-preview" }
		form = document.getElementById(where)
		form.insertBefore prevpost, form.firstChild

	prevpost.removeChild prevpost.firstChild while prevpost.firstChild?

	previewAuthor = window.createElementEx "h3", { className: "post-author" }
	previewAuthor.appendChild document.createTextNode "Post preview"
	prevpost.appendChild previewAuthor

	previewContent = window.createElementEx "div", { className: "post-content typebbcode" }
	previewContent.innerHTML = content
	prevpost.appendChild previewContent

# post edit
editPost = (id) ->
	if id == 0
		pid = type = "thread"
		idname = "OP"
	else
		pid = "p#{id}"
		type = "post"
		idname = "##{id}"
	post = document.getElementById pid
	nickspan = document.querySelector "##{pid} .nickname"
	nick = nickspan.innerHTML
	tripcodebar = null
	if !window.crOpts.adminMode
		if nickspan.parentNode.querySelector('span.tripcode')?.innerHTML.length > 0
			# visible tripcode
			tripcodebar = window.createElementEx "input", {
				type:        "text",
				name:        "tripcode",
				placeholder: "Tripcode (required)",
				required:    true,
				className:   "full short inline verysmall"
			}
		else
			# hidden tripcode (post-author contains <span class="anon"></span> instead of nick)
			htrip = JSON.parse(window.localStorage?.getItem 'crLatestPost')?.htrip
			tripcodebar = window.createElementEx "input", {
				type:     "text",
				name:     "tripcode",
				required: true,
				value:    htrip
			}

	# if post is OP, allow editing thread tags
	tagsbar = null
	tags = post.dataset?.tags
	if idname is "OP"
		tagsbar = window.createElementEx "input", {
			type:        "text",
			name:        "tags",
			placeholder: "Tags (separated by #)",
			className:   "full small",
			dataset:     { acsearch: false },
			value:       tags
		}
	original[id] = post.innerHTML

	# Editor
	section = window.createElementEx "section", { id: id, className: "form" }

	anchor = window.createElementEx "a", { className: "nolink", name: "edit" }
	section.appendChild anchor

	editform = window.createElementEx "form", { id: "edit#{id}", method: "POST" }
	editform.action = window.stripPage(location.pathname) + "/post/" + id + "/edit"
	section.appendChild editform

	nickbar = document.createElement "div"
	editform.appendChild nickbar

	nickname = window.createElementEx "span", {
		className: "full verysmall nickname",
		style: { display: "inline-block", border: "0", width: "auto" }
	}
	nickname.appendChild document.createTextNode nick
	nickbar.appendChild nickname

	if tripcodebar?
		nickbar.appendChild tripcodebar

	editingNo = window.createElementEx "span", {
		style: { display: "inline-block", color: "#ccc", fontSize: "0.9em", width: "auto" }
	}
	editingNo.appendChild document.createTextNode "editing #{idname}"
	nickbar.appendChild editingNo


	editorRight = window.createElementEx "div", { className: "small editorButtonCont editorButtonContRight" }
	editform.appendChild editorRight

	formattingHelper = window.createElementEx "a", { target: "_blank", href: "/stiki/formatting", rel: "help" }
	formattingHelper.appendChild document.createTextNode "?"
	editorRight.appendChild formattingHelper

	editorEditButtons = window.createElementEx "div", { className: "small editorButtonCont" }
	addEditorButtons editorEditButtons
	editform.appendChild editorEditButtons

	textarea = window.createElementEx "textarea", {
		className:   "full small editor",
		name:        "text",
		id:          "edit#{id}-text",
		required:    true,
		placeholder: "Retreiving content..."
	}
	editform.appendChild textarea

	if tagsbar?
		section.appendChild tagsbar

	centerButtonCont = window.createElementEx "div", { className: "center" }
	editform.appendChild centerButtonCont

	charCounter = window.createElementEx "div", { className: "chars-count" }
	charCounter.dataset["maxlen"] = window.crOpts.maxlen
	centerButtonCont.appendChild charCounter

	submit = window.createElementEx "input", { type: "submit", value: "Edit post" }
	centerButtonCont.appendChild submit

	cancel = window.createElementEx "button", { type: "button" }
	cancel.addEventListener "click", cancelForm id
	cancel.appendChild document.createTextNode "Cancel"
	centerButtonCont.appendChild cancel

	preview = window.createElementEx "input", { type: "button", className: "button", value: "Preview" }
	preview.addEventListener "click", showPreview "edit#{id}"
	centerButtonCont.appendChild preview

	post.removeChild post.firstChild while post.firstChild?
	post.appendChild section

	qwest.post(window.stripPage(location.pathname) + "/post/" + id + "/raw")
		.then (resp) ->
			textarea.value = resp
			textarea.placeholder = "Post content (Markdown, HTML and BBCode are supported)"
			charsCount "edit#{id}"
		.catch (err) ->
			return if section.firstChild.className == "errmsg"
			section = document.getElementById id
			errmsg = window.createElementEx 'p', {
				className: "errmsg"
			}
			errmsg.appendChild = document.createTextNode "Failed to retrieve content: #{err}"
			section.insertBefore errmsg, section.firstChild

	return

# post delete
deletePost = (id) ->
	if id == 0
		pid = type = "thread"
		idname = "OP"
	else
		pid = "p#{id}"
		type = "post"
		idname = "##{id}"
	post = document.getElementById pid
	nickspan = document.querySelector "##{pid} .nickname"
	nick = nickspan.innerHTML
	original[id] = post.innerHTML
	tripcodebar = ""
	if !window.crOpts.adminMode
		purge = ""
		if nickspan.parentNode.querySelector('span.tripcode')?.innerHTML.length > 0
			# visible tripcode
			tripcodebar = "<input class='full short inline verysmall' type='text' name='tripcode' placeholder='Tripcode (required)' required />"
		else
			# hidden tripcode
			htrip = JSON.parse(window.localStorage?.getItem 'crLatestPost')?.htrip
			tripcodebar = "<input type='hidden' name='tripcode' value='#{htrip}' required />"

	else
		purge = '<button name="deletetype" value="purge" type="submit">Purge</button>'
	post.innerHTML = """
<section id="#{id}" class="form"><a name="delete" class="noborder"></a>
    <form method="POST" action="#{window.stripPage(location.pathname) + "/post/" + id + "/delete"}">
        <div>
            <span class="full verysmall nickname" style="display: inline-block; border: 0; width: auto">#{nick}</span>
            #{tripcodebar}
            <span style="color: #ccc; display: inline-block; width: auto; font-size: 0.9em;">deleting ##{id}</span>
        </div>
        <div class="center">
            <button name="deletetype" value="soft" type="submit">Delete</button>#{purge}<button type="button" onclick="Posts.cancelForm(#{id});">Cancel</button>
        </div>
  </form>
</section>"""
	return

# check form before submitting
replyPreSubmit = (elem, threadUrl, curPostCount) ->
	nick = document.querySelector("##{elem.id} input[name='nickname']").value
	if nick.indexOf('#') > 0 and nick.indexOf('#') == nick.length - 1
		alert "Tripcode must have at least 1 character."
		return false
	return true

# post quote by id
quotePostId = (id) ->
	return ->
		text = document.getElementById "reply-form-text"
		if text.value.length > 0 and text.value[text.value.length - 1] isnt "\n"
			text.value += "\n>> ##{id}\n"
		else
			text.value += ">> ##{id}\n"
		text.focus()

# remove fallback and set onclick events
fromList(document.getElementsByClassName 'postEditLink').map (e) ->
	postId = e.dataset?.postid
	return unless postId?
	e.href = "#p#{postId}"
	e.className ="postEditLink edit nolink"
	do (postId) ->
		e.onclick = -> editPost parseInt(postId, 10)

fromList(document.getElementsByClassName 'postDeleteLink').map (e) ->
	postId = e.dataset?.postid
	return unless postId?
	e.href = "#p#{postId}"
	e.className ="postDeleteLink delete nolink"
	do (postId) ->
		e.onclick = -> deletePost parseInt(postId, 10)

preview = document.getElementById "preview-post"
if preview?
	preview.style.display = ""
	preview.addEventListener "click", showPreview "reply-form"

editorButtons = document.getElementById "editorButtons"
if editorButtons?
	addEditorButtons editorButtons

# Add quote buttons
fromList(document.getElementsByClassName 'postQuoteButton').map (e) ->
	e.addEventListener "click", quotePostId e.dataset.postid

# expose functions
window.Posts =
	cancelForm: cancelForm
	replyPreSubmit: replyPreSubmit
	quotePostId: quotePostId