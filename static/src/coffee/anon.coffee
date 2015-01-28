anons = ["Rock", "Boulder", "Pebble", "Mineral", "Stone"]

randomAnon = () -> anons[Math.floor(Math.random()*anons.length)]

fromList(document.querySelectorAll ".thread-author,.post-author").map (elem) ->
	nick = elem.querySelector ".nickname"
	trip = elem.querySelector ".tripcode"
	return if nick.innerHTML.length + trip.innerHTML.length > 0
	nick.innerHTML = "<span class=\"anon\">"+randomAnon()+"</span>"

# check if any post in this page is editable (i.e. we have the hidden tripcode
# for it in localStorage)
if window.localStorage?.getItem('crLatestPost')?
	latestPost = JSON.parse window.localStorage.getItem 'crLatestPost'
	return unless latestPost.thread == location.pathname.split('/')[2]
	btnDiv = document.getElementById "p#{latestPost.post}_btn"
	return unless btnDiv?
	# mark the post as editable
	btnDiv.style.display = "block"
