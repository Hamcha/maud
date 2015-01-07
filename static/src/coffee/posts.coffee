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
    content = "Retrieving content..."
    qwest.post(window.stripPage(location.pathname) + "/post/" + id + "/raw")
        .then (resp) ->
            content = resp
            document.querySelector("##{pid} textarea[name='text']").value = content
        .catch (err) ->
            content = ""
            section = document.getElementById(id)
            errmsg = document.createElement 'p'
            errmsg.className = "errmsg"
            errmsg.innerHTML = "Failed to retrieve content."
            section.insertBefore errmsg, section.firstChild
    nick = document.querySelector("##{pid}  .nickname").innerHTML
    tripcodebar = if !window.adminMode then "<input class=\"full short inline verysmall\" type=\"text\" name=\"tripcode\" placeholder=\"Tripcode (required)\" required />" else ""
    # if post is OP, allow editing thread tags
    tagsbar = ""
    tags = post.dataset?.tags
    if idname is "OP" and tags.replace(/// ///g, '').length > 0
        tagsbar = "<input class=\"full small\" type=\"text\" name=\"tags\" placeholder=\"Tags (separated by #)\" value=\"#{tags}\"/>"
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
    <div id="editorButtons" class="small">
        <a onclick="editorAdd(this, 'b')"><b>B</b></a>
        <a onclick="editorAdd(this, 'i')"><i>i</i></a>
        <a onclick="editorAdd(this, 'u')"><u>u</u></a>
        <a onclick="editorAdd(this, 'strike')"><strike>strike</strike></a>
        <a onclick="editorAdd(this, 'img')">img</a>
        <a onclick="editorAdd(this, 'url')"><span style="border-bottom: 1px dotted #fff">url</span></a>
        <a onclick="editorAdd(this, 'spoiler')">spoiler</a>
        <a onclick="editorAdd(this, 'youtube')">youtube</a>
    </div>
    <textarea class="full small editor" name="text" required placeholder="Thread text (Markdown is supported)">#{content}</textarea>
    #{tagsbar}
    <center>
      <div class="chars-count" data-maxlen="#{maxlen}"></div>
      <input type="Submit" value="Edit post"/><button type="button" onclick="cancelForm(#{id});">Cancel</button>
    </center>
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
    nick = document.querySelector("##{pid}  .nickname").innerHTML
    original[id] = post.innerHTML
    tripcodebar = if !window.adminMode then "<input class=\"full short inline verysmall\" type=\"text\" name=\"tripcode\" placeholder=\"Tripcode (required)\" required />" else ""
    post.innerHTML = """
<section id="#{id}" class="form"><a name="delete" class="noborder"></a>
  <form method="POST" action="#{window.stripPage(location.pathname) + "/post/" + id + "/delete"}">
    <div>
      <span class="full verysmall nickname" style="display: inline-block; border: 0; width: auto">#{nick}</span>
      #{tripcodebar}
      <span style="color: #ccc; display: inline-block; width: auto; font-size: 0.9em;">deleting ##{id}</span>
    </div>
    <center>
      <button type="submit">Delete post</button><button type="button" onclick="cancelForm(#{id});">Cancel</button>
    </center>
  </form>
</section>"""
    return

cancelForm = (id) ->
    pid = if id == 0 then "thread" else "p#{id}"
    post = document.getElementById pid
    post.innerHTML = original[id]

# post preview
showPreview = () ->
    form = document.getElementById 'prev-form'
    text = document.querySelector("#prev-form textarea[name='text']").value
    nick = document.querySelector("#prev-form input[name='nickname']").value
    unless text
        if form.firstChild.className == 'errmsg'
            return
        errmsg = document.createElement 'p'
        errmsg.className = 'errmsg'
        errmsg.innerHTML = "Please insert some content."
        form.insertBefore errmsg, form.firstChild
        return
    req =
        nickname: nick
        text:     text
    # retreive content data from the server
    qwest.post('/postpreview', req)
        .then (resp) ->
            console.log resp
            createPreview resp
        .catch (err) ->
            console.log 'Error!'
            console.log err

createPreview = (content) ->
    # deselect selected post, if any
    o.className = o.className.replace "post-selected", "" for o in document.querySelectorAll ".post-selected"
    # if preview post already exists, just update it
    prevpost = document.getElementById 'post-preview'
    unless prevpost
        prevpost = document.createElement 'article'
        prevpost.id = 'post-preview'
        # insert preview before the preview form
        document.getElementById('prev-form').parentNode.insertBefore prevpost, document.getElementById 'prev-form'
    prevpost.innerHTML = """<h3 class="post-author">Post preview</h3>
    <div class="post-content typebbcode">#{content}</div>
    """

window.editPost = editPost
window.deletePost = deletePost
window.cancelForm = cancelForm
window.showPreview = showPreview
