window.OptionUtils =
	showOptions: ->
		div = document.getElementById 'overlay'
		div.style.visibility = if div.style.visibility is 'visible' then 'hidden' else 'visible'
		unless window.localStorage?
			warn = document.createElement 'p'
			warn.innerHTML = "Your browser doesn't support localStorage: cannot change site option."
			insertAfter warn, div.querySelector('h2')
			div.querySelector('form').innerHTML = ""

	setOptions: (form) ->
		return unless window.localStorage?
		opts = {}
		fromList(document.getElementsByName 'option').map (opt) ->
			return unless opt.id?
			opts[opt.id] = opt.checked
		# save options in localStorage (it's no use changing the SiteOptions
		# live, because the new options will be effective only after a page reload).
		window.localStorage.setItem 'crOptions', JSON.stringify opts

# Load options
opts = window.localStorage?.getItem 'crOptions'
return unless opts?
window.SiteOptions = JSON.parse opts
fromList(document.getElementsByName 'option').map (opt) ->
	return unless opt.id?
	opt.checked = window.SiteOptions[opt.id]
