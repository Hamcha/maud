return unless window.localStorage?

if window.location.pathname[1...7] == 'thread'
	# When a thread is visited, save the latest post date is localStorage.
	# First, ensure this is the last page of the thread. If not, don't
	# mark this thread as 'visited', since latest replies are not being
	# actually seen.
	pages = document.querySelector 'div.pages'
	return unless pages.dataset.current == pages.dataset.max
	# grab latest post
	posts = document.getElementById('replies').querySelectorAll 'article.post'
	latest = posts[posts.length-1]
	date = latest.querySelector('a.date').dataset.udate
	if date?
		surl = window.location.pathname.split('/')[2]
		window.localStorage.setItem "lview_#{surl}", date
else
	# In home: for each thread, check if already viewed or not
	window.fromList(document.querySelectorAll 'article.thread').map (thread) ->
		# get last-modified date
		date = thread.querySelector('span.date').dataset.udate
		# compare with locally saved date, if any
		surl = thread.querySelector('a').pathname.split('/')[2]
		lview = window.localStorage.getItem "lview_#{surl}"
		if lview? and lview == date
			# no updates since latest visit
			thread.querySelector('.thread-name').style.opacity = 0.7
