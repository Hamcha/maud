window.fromList = (x) -> Array.prototype.slice.call x

window.stripPage = (url) ->
	idx = url.indexOf "/page/"
	return if idx < 0 then url else url.substring 0, idx

window.setFilter = (tags) ->
	tagList = tags.join ":"
	return false unless Cookies.enabled
	props =
		domain: '.' + window.domain.split(":")[0]
		explires: Infinity
	console.log props
	Cookies.set "filter", tagList, props
	return true

window.getFilter = (tags) ->
	return false unless Cookies.enabled
	tagList = Cookies.get "filter"
	return [] unless tagList?
	return tagList.split ":"

window.addFilter = (toadd) ->
	tags = window.getFilter()
	for tag in toadd
		tags.push tag
	window.setFilter tags

window.removeFilter = (toremove) ->
	tags = window.getFilter()
	newtags = []
	for tag in tags
		continue if tag in toremove
		newtags.push tag
	window.setFilter newtags