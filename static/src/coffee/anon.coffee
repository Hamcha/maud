anons = ["Rock", "Boulder", "Pebble", "Mineral", "Stone"]

randomAnon = () -> anons[Math.floor(Math.random()*anons.length)]

fromList(document.querySelectorAll ".thread-author,.post-author").map (elem) ->
	anon = elem.querySelector ".anon"
	return unless anon isnt null
	anon.innerHTML = randomAnon()

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
