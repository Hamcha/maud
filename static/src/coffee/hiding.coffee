safeButton = document.getElementById "safeBtn"
unless Cookies.enabled
	safeButton.onclick = -> alert "Cookies are not enabled, Safe mode not available."
	return

# Wrapper for the cookie used for hiding.
# This cookie's value has the form: 'surl1&surl2&#tag1&#tag2&etc'
class Hidden
	constructor: (@value, @sep = '&') -> @splitted = @value.split @sep
	tags: -> return (x for x in @splitted when x[0] is '#')
	threads: -> return (x for x in @splitted when x[0] isnt '#')
	remove: (what...) ->
		for w in what
			@splitted.splice @splitted.indexOf(w), 1
			@value = @splitted.join @sep
		@update()
	add: (what...) ->
		for w in what
			continue if w in @splitted
			@splitted.push w
		@value = @splitted.join @sep
		@update()
	clear: -> @value = ""; @splitted = []; @update()
	clearThreads: ->
		@splitted = @tags()
		@value = @splitted.join @sep
		@update()
	clearTags: ->
		@splitted = @threads()
		@value = @splitted.join @sep
		@update()
	get: (what) ->
		i = @splitted.indexOf what
		return null if i < 0
		return @splitted[i]
	isEmpty: -> return @value.length is 0
	update: ->
		if @isEmpty()
			Cookies.expire 'crHidden'
		else
			Cookies.set 'crHidden', @value, { expires: Infinity }

	@expire: -> Cookies.expire 'crHidden'

crHidden = undefined
# find out if the cookies is set and, if so, create a Hidden object
# wrapping its value.
cookie = Cookies.get 'crHidden'
crHidden = new Hidden cookie if cookie?

# ensure we don't carry around a stale empty cookie
Hidden.expire() if crHidden?.isEmpty()

# Setup Safe mode button
if crHidden?.get "#nsfw"
	safeButton.innerHTML = "EXIT SAFE MODE"
	safeButton.style.boxShadow = "0 0 0 1px green inset"
	safeButton.onclick = ->
		crHidden.remove '#nsfw'
		location.reload true
		return
else
	safeButton.style.boxShadow = "0 0 0 1px darkred inset"
	safeButton.onclick = ->
		if crHidden?
			crHidden.add '#nsfw'
		else
			crHidden = new Hidden '#nsfw'
			crHidden.update()
		location.reload true
		return

toggleHide = (elem) ->
	# check if this element is already hidden: if so,
	# unhide it, else hide it.
	if crHidden?.get elem
		crHidden.remove elem
	else
		if crHidden?
			crHidden.add elem
		else
			crHidden = new Hidden elem
			crHidden.update()
		location.reload true
		return

# Bind click events on threads and tags
fromList(document.querySelectorAll 'div.hiding').map (e) ->
        e.innerHTML = """<a class="noborder hide" href="" onclick="Hiding.toggleHide('#{e.dataset.arg}')">Hide</a>"""

window.Hiding =
	toggleHide: toggleHide
	unhideAllThreads: -> crHidden.clearThreads()
	unhideAllTags: -> crHidden.clearTags()
