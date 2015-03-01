# Setup Safe mode button
safeButton = document.getElementById "safeBtn"
filter = window.getFilter()
if "nsfw" in filter
	safeButton.innerHTML = "EXIT SAFE MODE"
	safeButton.style.boxShadow = "0 0 0 1px green inset"
	safeButton.onclick = ->
		status = window.removeFilter ["nsfw"]
		location.reload true
		return
else
	safeButton.style.boxShadow = "0 0 0 1px darkred inset"
	safeButton.onclick = ->
		status = window.addFilter ["nsfw"]
		if status == false
			alert "Cookies are not enabled, Safe mode couldn't be enabled"
		location.reload true
		return

# Hiding functions
toggleHideThread = (url) ->
	# check if crHidden cookie exists
	cookie = Cookies.get 'crHidden'
	unless cookie?
		Cookies.set 'crHidden', url, { expires: Infinity }
		return
	hidden = cookie.split ' '
	if url in hidden
		Cookies.set 'crHidden', ((u for u in hidden when u isnt url).join ' '), { expires: Infinity }
	else
		Cookies.set 'crHidden', "#{cookie} #{url}", { expires: Infinity }

window.toggleHideThread = toggleHideThread

showHiddenThreads = ->
	
