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
    qwest.post(location.pathname + "/post/" + id + "/raw")
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
    original[id] = post.innerHTML
    post.innerHTML = """
<section id="#{id}" class="form"><a name="edit" class="nolink"></a>
  <form method="POST" action="#{location.pathname + "/post/" + id + "/edit"}">
    <div>
      <span class="full verysmall nickname" style="display: inline-block; border: 0; width: auto">#{nick}</span>
      <input class="full short inline verysmall" type="text" name="tripcode" placeholder="Tripcode (required)" required />
      <span style="color: #ccc; display: inline-block; width: auto; font-size: 0.9em;">editing #{idname}</span>
    </div>
    <textarea class="full small editor" name="text" required placeholder="Thread text (Markdown is supported)">#{content}</textarea>
    <center>
      <button onclick="editAJAX(#{id})">Edit post</button><button type="button" onclick="cancelForm(#{id});">Cancel</button>
    </center>
  </form>
</section>"""
    return

editAJAX = (id) ->
    qwest.post(location.pathname + "/post/" + id + "/edit")
        .then (resp) ->
            cancelForm(id)
            document.querySelector("#p#{id} textarea[name='text']").value = resp
        .catch (err) ->
            section = document.getElementById id
            errmsg = document.createElement 'p'
            errmsg.className = "errmsg"
            errmsg.innerHTML = "An error occurred: #{err}"
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
    nick = document.querySelector("##{pid}  .nickname").innerHTML
    original[id] = post.innerHTML
    post.innerHTML = """
<section id="#{id}" class="form"><a name="delete" class="noborder"></a>
  <form method="POST" action="#{location.pathname + "/post/" + id + "/delete"}">
    <div>
      <span class="full verysmall nickname" style="display: inline-block; border: 0; width: auto">#{nick}</span>
      <input class="full short inline verysmall" type="text" name="tripcode" placeholder="Tripcode (required)" required />
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
    form = document.getElementById 'reply-form'
    text = document.querySelector("#reply-form textarea[name='text']").value
    nick = document.querySelector("#reply-form input[name='nickname']").value
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
    # if preview post already exists, just update it
    prevpost = document.getElementById 'post-preview'
    unless prevpost
        prevpost = document.createElement 'article'
        prevpost.id = 'post-preview'
        document.getElementById('replies').appendChild prevpost
    prevpost.innerHTML = """<h4><em>Post preview</em></h4>
    <div class="post-content typebbcode">#{content}</div>
    """

window.editPost = editPost
window.deletePost = deletePost
window.cancelForm = cancelForm
window.showPreview = showPreview
