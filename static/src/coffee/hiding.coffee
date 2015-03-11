safeButton = document.getElementById "safeBtn"
unless Cookies.enabled
	safeButton.onclick = -> alert "Cookies are not enabled, Safe mode not available."
	return

# Wrapper for the cookie used for hiding.
# This cookie's value has the form: 'surl1&surl2&#tag1&#tag2&etc'
class Hidden
	sep: '&'
	constructor: (@value) -> @splitted = @value.split @sep
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

window.newCrHidden = ->
	cookie = Cookies.set 'crHidden', ""
	return new Hidden ""

window.killCrHidden = -> Cookies.expire 'crHidden'

crHidden = undefined
# find out if the cookies is set and, if so, create a Hidden object
# wrapping its value.
cookie = Cookies.get 'crHidden'
if cookie?
	crHidden = new Hidden cookie

# Setup Safe mode button
if crHidden?.get '#nfsw'
	safeButton.innerHTML = "EXIT SAFE MODE"
	safeButton.style.boxShadow = "0 0 0 1px green inset"
	safeButton.onclick = ->
		crHidden.remove '#nfsw'
		location.reload true
		return
else
	safeButton.style.boxShadow = "0 0 0 1px darkred inset"
	safeButton.onclick = ->
		crHidden = newCrHidden() unless crHidden?
		crHidden.add '#nfsw'
		location.reload true
		return

toggleHideThread = (url) ->
	# check if this url is already hidden: if so,
	# unhide it, else hide it.
	if crHidden?.get url
		crHidden.remove url
	else
		crHidden = newCrHidden() unless crHidden?
		crHidden.add url

window.toggleHideThread = toggleHideThread
window.unhideAllThreads = -> crHidden.clearThreads()
window.unhideAllTags = -> crHidden.clearTags()
