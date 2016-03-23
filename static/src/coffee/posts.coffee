'use strict'

original = []

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
	qwest.post(window.stripPage(location.pathname) + "/post/" + id + "/raw")
		.then (resp) ->
			textarea = document.querySelector("##{pid} textarea[name='text']")
			textarea.value = resp
			textarea.placeholder = "Post content (Markdown, HTML and BBCode are supported)"
		.catch (err) ->
			return if section.firstChild.className == "errmsg"
			section = document.getElementById id
			errmsg = document.createElement 'p'
			errmsg.className = "errmsg"
			errmsg.innerHTML = "Failed to retrieve content: #{err}"
			section.insertBefore errmsg, section.firstChild
	nickspan = document.querySelector "##{pid} .nickname"
	nick = nickspan.innerHTML
	tripcodebar = ""
	if !window.adminMode
		if nickspan.parentNode.querySelector('span.tripcode')?.innerHTML.length > 0
			# visible tripcode
			tripcodebar = "<input class='full short inline verysmall' type='text' name='tripcode' placeholder='Tripcode (required)' required />"
		else
			# hidden tripcode (post-author contains <span class="anon"></span> instead of nick)
			htrip = JSON.parse(window.localStorage?.getItem 'crLatestPost')?.htrip
			tripcodebar = "<input type='hidden' name='tripcode' value='#{htrip}' required />"

	# if post is OP, allow editing thread tags
	tagsbar = ""
	tags = post.dataset?.tags
	if idname is "OP"
		tagsbar = "<input class='full small' type='text' name='tags' placeholder='Tags (separated by #)' value='#{tags}'/>"
	original[id] = post.innerHTML
	post.innerHTML = """
<section id="#{id}" class="form"><a name="edit" class="nolink"></a>
    <form id="edit#{id}" method="POST" action="#{window.stripPage(location.pathname) + "/post/" + id + "/edit"}">
        <div>
            <span class="full verysmall nickname" style="display: inline-block; border: 0; width: auto">#{nick}</span>
            #{tripcodebar}
            <span style="color: #ccc; display: inline-block; width: auto; font-size: 0.9em;">editing #{idname}</span>
        </div>
        <!-- Editor buttons -->
        <div id="editorRight" class="small">
            <a target="_blank" href="/stiki/formatting" rel="help">?</a>
        </div>
        <div id="editorButtons" class="small">
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
        </div>
        <textarea class="full small editor" name="text" required placeholder="Retreiving content..."></textarea>
        #{tagsbar}
        <div class="center">
            <div class="chars-count" data-maxlen="#{maxlen}"></div>
            <input type="Submit" value="Edit post"/><button type="button" onclick="Posts.cancelForm(#{id});">Cancel</button>
            <input type="button" class="button" onclick="Posts.showPreview('edit#{id}')" value="Preview" />
        </div>
    </form>
</section>"""
	charsCount "edit#{id}"
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
	if !window.adminMode
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

cancelForm = (id) ->
	pid = if id == 0 then "thread" else "p#{id}"
	post = document.getElementById pid
	post.innerHTML = original[id]
	post.querySelector('.postEditLink')?.onclick = -> editPost +id
	post.querySelector('.postDeleteLink')?.onclick = -> deletePost +id

# post preview
showPreview = (where) ->
	form = document.getElementById where
	text = document.querySelector("##{where} textarea[name='text']").value
	unless text
		return if form.firstChild.className == 'errmsg'
		errmsg = document.createElement 'p'
		errmsg.className = 'errmsg'
		errmsg.innerHTML = "Please insert some content."
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
	unless prevpost
		prevpost = document.createElement 'article'
		prevpost.id = "#{where}-preview"
		# insert preview before the reply form
		document.getElementById(where).parentNode.insertBefore prevpost, document.getElementById where
	prevpost.innerHTML = """<h3 class="post-author">Post preview</h3>
	<div class="post-content typebbcode">#{content}</div>
	"""

# check form before submitting
replyPreSubmit = (elem, threadUrl, curPostCount) ->
	nick = document.querySelector("##{elem.id} input[name='nickname']").value
	if nick.indexOf('#') > 0 and nick.indexOf('#') == nick.length - 1
		alert "Tripcode must have at least 1 character."
		return false
	return true

# post quote by id
quotePostId = (id) ->
	text = document.querySelector "#reply-form textarea[name='text']"
	if text.value.length > 0 and text.value[text.value.length - 1] isnt "\n"
		text.value += "\n>> ##{id}\n"
	else
		text.value += ">> ##{id}\n"
	window.location.href = '#reply'
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

# expose functions
window.Posts =
	cancelForm: cancelForm
	showPreview: showPreview
	replyPreSubmit: replyPreSubmit
	quotePostId: quotePostId
