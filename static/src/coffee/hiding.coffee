safeButton = document.getElementById "safeBtn"
unless Cookies.enabled
	safeButton.onclick = -> alert "Cookies are not enabled, Safe mode not available."
	return

# Wrapper for the cookie used for hiding.
# This cookie's value has the form: 'surl1&surl2&#tag1&#tag2&etc'
Hidden =
	# Sets the new value of the hidden cookie
	new: (@value, @sep = '&') ->
		@splitted = @value.split @sep
		this

	tags: -> x for x in @splitted when x[0] is '#'
	
	threads: -> x for x in @splitted when x[0] isnt '#'
	
	# Removes a set of elements from the hidden cookie
	remove: (what...) ->
		for w in what
			@splitted.splice @splitted.indexOf(w), 1
			@value = @splitted.join @sep
		@update()
	
	# Adds a set of elements to the hidden cookie
	add: (what...) ->
		for w in what
			continue if w in @splitted
			@splitted.push w
		@value = @splitted.join @sep
		@update()
	
	# Resets the hidden cookie
	clear: ->
		@value = ""
		@splitted = []
		@update()
	
	# Removes all threads from the hidden cookie
	clearThreads: ->
		@splitted = @tags()
		@value = @splitted.join @sep
		@update()
	
	# Removes all tags from the hidden cookie
	clearTags: ->
		@splitted = @threads()
		@value = @splitted.join @sep
		@update()
	
	# Returns the index of the hidden element `what`, or null
	get: (what) ->
		i = @splitted.indexOf what
		return null if i < 0
		@splitted[i]
	
	isEmpty: -> @value.length is 0
	
	# Syncs the actual cookie with this object
	update: ->
		if @isEmpty()
			@expire()
		else
			Cookies.set 'crHidden', @value, { expires: Infinity }
	
	expire: -> Cookies.expire 'crHidden'

# find out if the cookies is set and, if so, wrap it
cookie = Cookies.get 'crHidden'
crHidden = cookie? && Hidden.new(cookie) || undefined

# ensure we don't carry around a stale empty cookie
crHidden?.update()

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
			crHidden = Hidden.new '#nsfw'
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
			crHidden = Hidden.new elem
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
