'use strict'

showOptions = ->
	div = document.getElementById 'overlay'
	div.style.visibility = if div.style.visibility is 'visible' then 'hidden' else 'visible'
	unless window.localStorage?
		warn = document.createElement 'p'
		warn.innerHTML = "Your browser doesn't support localStorage: cannot change site option."
		insertAfter warn, div.querySelector('h2')
		div.querySelector('form').innerHTML = ""

reloadOptions = ->
	return unless window.localStorage?
	opts = {}
	fromList(document.getElementsByName 'option').map (opt) ->
		return unless opt.id?
		opts[opt.id] = opt.checked
	# save options in localStorage (it's no use changing the SiteOptions
	# live, because the new options will be effective only after a page reload).
	window.localStorage.setItem 'crOptions', JSON.stringify opts
	toggleImgProxy()

toggleImgProxy = ->
	if document.getElementById('useProxy')?.checked
		Cookies.set 'crUseProxy', true, { expires: Infinity }
	else
		Cookies.expire 'crUseProxy'

# Load options
opts = window.localStorage?.getItem 'crOptions'
if opts?
	window.SiteOptions = JSON.parse opts
	fromList(document.getElementsByName 'option').map (opt) ->
		return unless opt.id?
		opt.checked = window.SiteOptions[opt.id]
	# Enable lock-tick if using proxy
	if Cookies.get('crUseProxy')
		sign = document.getElementById('secureSign')
		if location.protocol isnt 'https:'
			img = sign.querySelector('img')
			img.src = img.src[..-8] + "nosec.svg"
			img.alt = 'Enchanced Security requested, but protocol is not HTTPS'
		sign.style.visibility = 'visible'
else
	# likely the first time visiting the site
	window.SiteOptions = reloadOptions()

# Bind events
document.getElementById('siteOptionsShow').addEventListener 'click', showOptions
document.getElementById('siteOptionsHide').addEventListener 'click', showOptions
document.querySelector('#siteOptions form').addEventListener 'change', reloadOptions
