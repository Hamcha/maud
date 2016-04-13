'use strict'

###
  AJAX tags autocomplete plugin (requires qwest)
  How to use:
     1. Create a wrapper element (e.g. a form) with an ID
     2. Inside the wrapper, there must be an element with class 'ac_input'
        (most likely an input)
     3. Call toggleAutocomplete(the_ac_input_element, '/url_where_to_retreive_data'[, {opts}])
  The ac_input element supports options given via data-*; currently implemented:
     - data-ac_search: if "on", then the ac list will have the 'search»' quick link, which
                       updates the form and submits it automatically.
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
	search = if elem.dataset?.acsearch == 'on' then on else off
	elem.addEventListener "keyup", (e) ->
		curTag =
			if elem.value.indexOf sep > 0
				elem.value[elem.value.lastIndexOf(sep) + 1..].trim()
			else
				elem.value.trim()
		if not opts?.minChars? or curTag.length >= opts.minChars
			updateAutocompleteList ul, curTag, data, { search: search }
		else
			ul.innerHTML = ""
		ul.style.width = "#{elem.offsetWidth}px"
		ul.style.top = "#{elem.offsetTop + elem.offsetHeight}px"
		ul.style.left = "#{elem.offsetLeft}px"
		ul.style.visibility = if ul.innerHTML.length > 0 then 'visible' else 'hidden'

updateTags = (form, tag, list) ->
	# input to append the tags to
	el = form.querySelector '.ac_input'
	v = el.value
	if v.lastIndexOf(sep) > 0
		el.value = v[0..v.lastIndexOf(sep)-1] + "#{sep}#{tag} #"
	else
		el.value = "#{sep}#{tag} #"
	el.selectionStart = el.value.length
	el.focus()
	list.style.visibility = 'hidden'

updateAutocompleteList = (list, txt, data, opts = {}) ->
	count = 0
	searchanchor = null
	prevent = false
	if opts?.search isnt off
		searchanchor = (form, tag, list) ->
			a = window.createElementEx "a", { href: "#", className: "noborder rightanchor" }
			a.addEventListener "click", ->
				updateTags form, tag, list
				prevent = true
				form = list.parentNode
				form.onsubmit() if form.onsubmit?
				form.submit()
			a.appendChild document.createTextNode "search »"
			return a

	list.removeChild list.firstChild while list.firstChild?
	for el in data when el.trim().length > 0
		if el[0..txt.length-1] == txt and count++ < limit
			li = window.createElementEx "li", { title: el, style: { cursor: "pointer" } }
			li.addEventListener "click", ->
				return if prevent
				updateTags list.parentNode, el, list
			span = window.createElementEx "span"
			span.appendChild document.createTextNode el
			li.appendChild span
			if searchanchor?
				li.appendChild searchanchor list.parentNode, el, list
			list.appendChild li

window.AC =
	toggleAutocomplete: toggleAutocomplete