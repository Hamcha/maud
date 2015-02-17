###
  AJAX tags autocomplete plugin
  (requires qwest)
###

# autocomplete input from a JSON list (retreived via AJAX)
# opts:
#   - minChars
toggleAutocomplete = (elem, url, opts) ->
	console.log "elem: #{elem}"
	return unless elem?
	# get the JSON from the server
	data = []
	qwest.post(url, null, { responseType: 'json', async: false })
		.then (resp) ->
			data = resp
		.catch (err) ->
			console.log 'Error retreiving data'
	# element holding the autocomplete data
	ul = document.createElement 'ul'
	ul.className = 'autocomplete-list'
	insertAfter = (newNode, node) ->
		node.parentNode.insertBefore(newNode, node.nextSibling)
	insertAfter ul, elem
	console.log "data: #{data}"
	elem.onkeyup = (e) ->
		curTag = if elem.value.indexOf ',' > 0 then elem.value[elem.value.lastIndexOf(',') + 1..].trim() else elem.value.trim()
		if !opts.minChars? || curTag.length >= opts.minChars
			updateAutocompleteList ul, curTag, data
		else
			ul.innerHTML = ""

updateAutocompleteList = (list, txt, data) ->
	console.log "called autocomplete. txt: #{txt}, data: #{data}"
	list.innerHTML = ("<li><a class='noborder' href='#' onclick='acUpdateTags(\"#{list.parentNode.firstChild.id}\",\"#{el}\")'>#{el}</a></li>" for el in data when el.startsWith txt).join "\n"

#autocomplete '#tagsearch', '/taglist', { minChars: 2 }
# TODO: integrate with Taggle
#autocomplete '.taggle_input', '/taglist', {
#	minChars: 2,
#	keyupFunc: (key) ->
#		if key.keyCode is 188 # comma
#			newThreadTaggle.add(document.querySelector('.taggle_input').value)
#			return true
#		return false
#}

window.toggleAutocomplete = toggleAutocomplete
# FIXME
window.acUpdateTags = (elemId, tag) ->
	el = document.querySelector "##{elemId} ul"
	v = el.value
	if v.indexOf ',' > 0
		el.value = v[0..v.lastIndexOf ','] + ', ' + tag
	else
		el.value = tag
