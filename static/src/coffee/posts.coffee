original = []

# utilities
cancelForm = (id) ->
    post = document.getElementById "p#{id}"
    post.innerHTML = original[id]

# post edit
editPost = (id) ->
    post = document.getElementById "p#{id}"
    nick = content = null
    traverse = (p) ->
        for c in p.children
            traverse c if c.children.length > 0
            switch
              when c.className?.match /post-content/ then content = c.innerHTML
              when c.className?.match /nickname/ then nick = c.innerHTML
        return
    traverse post
    original[id] = post.innerHTML
    post.innerHTML = """
<section id="#{id}" class="form"><a name="edit" class="noborder"></a>
  <form method="POST" action="#{location.pathname + "/post/" + id + "/edit"}">
    <div>
      <span class="full verysmall nickname" style="display: inline-block; border: 0; width: auto">#{nick}</span>
      <input class="full short inline verysmall" type="text" name="tripcode" placeholder="Tripcode (required)" />
      <span style="color: #ccc; display: inline-block; width: auto; font-size: 0.9em;">editing ##{id}</span>
    </div>
    <textarea class="full verysmall" name="text" required placeholder="Thread text (Markdown is supported)">#{content}</textarea>
    <center>
      <button type="submit">Edit post</button><button type="button" onclick="cancelForm(#{id});">Cancel</button>
    </center>
  </form>
</section>"""
    return

# post delete
deletePost = (id) ->
    post = document.getElementById "p#{id}"
    nick = null
    traverse = (p) ->
        for c in p.children
            traverse c if c.children.length > 0
            switch
              when c.className?.match /nickname/ then nick = c.innerHTML
        return
    traverse post
    original[id] = post.innerHTML
    post.innerHTML = """
<section id="#{id}" class="form"><a name="delete" class="noborder"></a>
  <form method="POST" action="#{location.pathname + "/post/" + id + "/delete"}">
    <div>
      <span class="full verysmall nickname" style="display: inline-block; border: 0; width: auto">#{nick}</span>
      <input class="full short inline verysmall" type="text" name="tripcode" placeholder="Tripcode (required)" />
      <span style="color: #ccc; display: inline-block; width: auto; font-size: 0.9em;">deleting ##{id}</span>
    </div>
    <center>
      <button type="submit">Delete post</button><button type="button" onclick="cancelForm(#{id});">Cancel</button>
    </center>
  </form>
</section>"""
    return

window.editPost = editPost
window.deletePost = deletePost
window.cancelForm = cancelForm
