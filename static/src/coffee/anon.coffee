anons = ["Rock", "Boulder", "Pebble", "Mineral", "Stone"]

randomAnon = -> anons[Math.floor(Math.random()*anons.length)]

fromList(document.querySelectorAll ".nickname").map (elem) ->
	nick = elem.innerHTML
	if nick instanceof HTMLSpanElement
		nick.innerHTML = randomAnon()
	else if nick.length > 0 and SiteOptions?.anonizeAll
		elem.innerHTML =
			if window.location.pathname[0...8] == '/thread/'
				"<span class='anon'>#{randomAnon()}</span>"
			else
				 ''
		elem.parentElement.querySelector(".tripcode").innerHTML = ''

# if 'crSetLatestPost' cookie is set, save hidden tripcode and delete the cookie
saveHiddenTripcode = (thread, post, htrip) ->
	storage = window.localStorage
	return unless storage?
	# save hidden tripcode in storage
	storage.setItem "crLatestPost", JSON.stringify { thread: thread, post: post, htrip: htrip }
	return

lpCookie = Cookies.get 'crSetLatestPost'
if lpCookie?
	[thread, post, htrip] = lpCookie.split '/'
	if thread? and post? and htrip?
		saveHiddenTripcode thread, post, htrip
	Cookies.expire 'crSetLatestPost', { path: '/thread/' }

# check if any post in this page is editable (i.e. we have the hidden tripcode
# for it in localStorage)
if window.localStorage?.getItem('crLatestPost')?
	latestPost = JSON.parse window.localStorage.getItem 'crLatestPost'
	return unless latestPost.thread == location.pathname.split('/')[2]
	btnDiv = document.getElementById "p#{latestPost.post}_btn"
	return unless btnDiv?
	# check if the post is deleted
	return if (document.querySelector "#p#{latestPost.post} .typedeleted")
	# mark the post as editable
	btnDiv.style.display = "block"
