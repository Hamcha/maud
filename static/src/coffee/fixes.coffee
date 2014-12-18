# <img> fixes
fromList(document.querySelectorAll "img").map (e) ->
    # ALT fix - Set titles to alt text (xkcd style)
    e.title = e.alt if e.alt != ""

# Handle hash changes
window.onhashchange = () ->
    return unless location.hash.length > 0
    # Post selected
    if location.hash[1] is 'p'
        o.className = o.className.replace "post-selected", "" for o in document.querySelectorAll ".post-selected"
        doc = document.querySelector location.hash
        doc.className = "post post-selected" if doc?
        return
window.onhashchange()

# Fix greentext on old posts
#TODO: move this to server parsing
fromList(document.querySelectorAll ".type blockquote p").map (e) ->
    e.innerHTML = "> " + e.innerHTML.split("\n").join "<br />> "

# Make page lists
pageDiv = document.getElementById "pages"
if pageDiv?
    page = parseInt pageDiv.getAttribute "current"
    baseurl = stripPage location.pathname
    maxstr = pageDiv.getAttribute "max"
    more = pageDiv.getAttribute "more"
    # Do Next/Previous only when we don't know the number of pages
    if maxstr == "nomax"
        pageHTML = ""
        if page > 1
            pageHTML += "<a href=\"#{baseurl}/page/#{page - 1}\">&laquo; Back</a> "
        pageHTML += "<b>#{page}</b>"
        if more == "yes"
            pageHTML += " <a href=\"#{baseurl}/page/#{page + 1}\">Next &raquo;</a>"
        pageDiv.innerHTML = pageHTML
    else
        max = parseInt maxstr
        if max > 1
            pageHTML = "PAGE &nbsp;"
            pageHTML += (if page == n then "<b>#{n}</b> " else "<a href=\"#{baseurl}/page/#{n}\">#{n}</a> ") for n in [1..max]
            pageDiv.innerHTML = pageHTML

do charsCount = ->
    form = document.getElementById 'prev-form'
    if form?
        text = document.querySelector("#prev-form textarea")
        span = document.querySelector("#prev-form .chars-count")
        text.onkeyup = () ->
            span.innerHTML = "#{span.dataset.maxlen - text.value.length} characters left"