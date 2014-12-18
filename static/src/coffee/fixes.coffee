# <img> fixes
fromList(document.querySelectorAll "img").map (e) ->
    # ALT fix - Set titles to alt text (xkcd style)
    e.title = e.alt if e.alt != ""

# Handle hash changes
window.onhashchange = () ->
    return unless location.hash.length > 0
    # Post selected
    if location.hash is "#last"
        o.className = o.className.replace "post-selected", "" for o in document.querySelectorAll ".post-selected"
        doc = document.querySelectorAll ".post"
        history.replaceState {}, document.title, location.pathname + "#" + doc[doc.length - 1].id
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
    max = parseInt pageDiv.getAttribute "max"
    if max > 1
        pageHTML = "PAGE &nbsp;"
        baseurl = stripPage location.pathname
        pageHTML += (if page == n then "<b>#{n}</b> " else "<a href=\"#{baseurl}/page/#{n}\">#{n}</a> ") for n in [1..max]
        pageDiv.innerHTML = pageHTML
