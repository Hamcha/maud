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
    content = document.querySelector("##{pid} ." + type + "-content").innerHTML
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
      <button type="submit">Edit post</button><button type="button" onclick="cancelForm(#{id});">Cancel</button>
    </center>
  </form>
</section>"""
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
    #form = document.getElementById 'reply-form'
    text = document.querySelector("#reply-form textarea").value
    nick = document.getElementById('reply-nickname').value
    unless text
        # TODO
        console.log 'no data'
        return
    req =
        nickname: nick
        text:     text
    # retreive content data from server
    qwest.post('/postpreview', req)
        .then (resp) ->
            console.log resp
            createPreview resp, nick
        .catch (err) ->
            console.log 'Error!'
            console.log err

createPreview = (content, nick) ->
    # if preview post already exists, just update it
    prevpost = document.getElementById 'post-preview'
    unless prevpost
        prevpost = document.createElement 'article'
    prevpost.innerHTML = """
    <h3 class="post-author">
""" + (nick ? """
        <span class="nickname">#{nick}</span>
""" : "") + """ 
    </h3>
    <div class="post-content type{{ContentType}}">{{{Content}}}</div>
    """

window.editPost = editPost
window.deletePost = deletePost
window.cancelForm = cancelForm
window.showPreview = showPreview
