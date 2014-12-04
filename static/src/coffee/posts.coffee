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
      <input class="full short inline verysmall" type="text" name="tripcode" placeholder="Tripcode (required)" />
      <span style="color: #ccc; display: inline-block; width: auto; font-size: 0.9em;">editing #{idname}</span>
    </div>
    <textarea class="full small editor" name="text" required placeholder="Thread text (Markdown is supported)">#{content}</textarea>
    <center>
      <button type="submit">Edit post</button><button type="button" onclick="cancelEdit(#{id});">Cancel</button>
    </center>
  </form>
</section>"""
    return

cancelEdit = (id) ->
    pid = if id == 0 then "thread" else "p#{id}"
    post = document.getElementById pid
    post.innerHTML = original[id]

window.editPost = editPost
window.cancelEdit = cancelEdit
