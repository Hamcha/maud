###
  AJAX tags autocomplete plugin (requires qwest)
  How to use:
     1. Create a wrapper element (e.g. a form) with an ID
     2. Inside the wrapper, there must be an element with class 'ac_input'
        (most likely an input)
     3. Call toggleAutocomplete(the_ac_input_element, '/url_where_to_retreive_data'[, {opts}])
###

# tag separator
sep = '#'

# autocomplete input from a JSON list (retreived via AJAX)
# opts:
#   - minChars
toggleAutocomplete = (elem, url, opts) ->
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
	ul.className = 'ac_list'
	ul.style.visibility = 'hidden'
	ul.style.zIndex = 10
	ul.id = 'ac_list'
	insertAfter = (newNode, node) ->
		node.parentNode.insertBefore(newNode, node.nextSibling)
	insertAfter ul, elem
	elem.onkeyup = (e) ->
		curTag =
			if elem.value.indexOf sep > 0
				elem.value[elem.value.lastIndexOf(sep) + 1..].trim()
			else
				elem.value.trim()
		if not opts?.minChars? or curTag.length >= opts.minChars
			updateAutocompleteList ul, curTag, data
		else
			ul.innerHTML = ""
		ul.style.visibility = if ul.innerHTML.length > 0 then 'visible' else 'hidden'

updateAutocompleteList = (list, txt, data) ->
	list.innerHTML =
		(for el in data
			if el[0..txt.length-1] == txt
				"<li><a class='noborder' href='#' onclick='acUpdateTags(" +
				"\"#{list.parentNode.id}\", \"#{el}\", " +
				"\"#{list.id}\")'>#{el}</a></li>"
		).join("\n").trim()


window.toggleAutocomplete = toggleAutocomplete

window.acUpdateTags = (formId, tag, listId) ->
	# input to append the tags to
	el = document.getElementById(formId).querySelector '.ac_input'
	v = el.value
	if v.lastIndexOf(sep) > 0
		el.value = v[0..v.lastIndexOf(sep)-1] + "#{sep}#{tag} "
	else
		el.value = "#{sep}#{tag} "
	el.selectionStart = el.value.length
	el.focus()
	document.getElementById(listId).style.visibility = 'hidden'
