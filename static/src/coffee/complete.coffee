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
# max amount of displayed tags
limit = 10

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
		ul.style.width = "#{elem.offsetWidth}px"
		ul.style.top = "#{elem.offsetTop + elem.offsetHeight}px"
		ul.style.left = "#{elem.offsetLeft}px"
		ul.style.visibility = if ul.innerHTML.length > 0 then 'visible' else 'hidden'

updateAutocompleteList = (list, txt, data) ->
	count = 0
	list.innerHTML =
		(for el in data
			if el[0..txt.length-1] == txt and count++ < limit
				utargs = """ "#{list.parentNode.id}","#{el}","#{list.id}" """
				"""
				<li title='#{el}' style='cursor:pointer' onclick='AC.prevent || AC.updateTags(#{utargs})'>
				    <span>#{el}</span>
				    <a href='#' class='noborder rightanchor' onclick='(function () {
					       AC.updateTags(#{utargs});
					       AC.prevent = true; // this is a quite ugly hack
					       document.getElementById("#{list.parentNode.id}").submit();
				       })()'>
				           search &raquo;
				    </a>
				</li>
				"""
		).join("\n").trim()


# expose AutoComplete functions
window.AC =
	toggleAutocomplete: toggleAutocomplete

	updateTags: (formId, tag, listId) ->
		# input to append the tags to
		el = document.getElementById(formId).querySelector '.ac_input'
		v = el.value
		if v.lastIndexOf(sep) > 0
			el.value = v[0..v.lastIndexOf(sep)-1] + "#{sep}#{tag} #"
		else
			el.value = "#{sep}#{tag} #"
		el.selectionStart = el.value.length
		el.focus()
		document.getElementById(listId).style.visibility = 'hidden'
