'use strict'

return unless window.localStorage?

stripBasepath = (s) -> s[window.crOpts.basepath.length..]

markAllRead = ->
	window.fromList(document.querySelectorAll 'article.thread-item').map (thread) ->
		date = thread.querySelector('span.date').dataset.udate
		lreplyAnchor = thread.querySelector 'a.last-reply'
		[surl, _, lpage] = lreplyAnchor.pathname.split('/')[2..4]
		lpost = lreplyAnchor.hash
		if lpost.length < 3
			# shouldn't happen, but for insurance we put post = 0
			lpost = 0
		else
			lpost = lpost[2..] # strip the initial "#p"
		lpage = '1' unless lpage?
		console.log "saving lview_#{surl}"
		window.localStorage.setItem "lview_#{surl}", "#{date}##{lpost}##{lpage}"

pageIs =
	thread:    window.location.pathname.startsWith "#{window.crOpts.basepath}thread/"
	threads:   window.location.pathname.startsWith "#{window.crOpts.basepath}threads"
	tagSearch: window.location.pathname.startsWith "#{window.crOpts.basepath}tag/"
	home:      window.location.pathname == window.crOpts.basepath

# mark NSFW threads
if pageIs.tagSearch or pageIs.threads or pageIs.home
	window.fromList(document.querySelectorAll 'article.thread-item, article.home-thread').map (thread) ->
		tags = thread.dataset?.tags?.split '#'
		return unless tags?
		# setup title
		thread.title = "#{tags.join ' #'}"[1..]
		if 'nsfw' in tags
			title = thread.querySelector('.thread-name, .thread-title')
			img = window.createElementEx "img", { src: "/static/images/nsfw.svg", alt: "NSFW", className: "nsfwIcon" }
			title.insertBefore img, title.firstChild


return unless SiteOptions?.dehighlight or SiteOptions?.jumptolastread

if pageIs.thread
	# When a thread is visited, save the latest post date is localStorage.
	# First, ensure this is the last page of the thread. If not, don't
	# mark this thread as 'visited', since latest replies are not being
	# actually seen. Also save number of replies, and make the thread link
	# point to the latest read post instead of last one.
	pages = document.querySelector 'div.pages'
	return unless pages?.dataset? and (pages.dataset.current == pages.dataset.max)
	# grab latest post
	posts = document.getElementById('replies').querySelectorAll 'article.post'
	nreplies = posts.length
	# if no replies, pick op
	latest = document.getElementById 'thread'
	npost = 0
	if posts.length > 0
		latest = posts[posts.length-1]
		npost = latest.id[1..]
	# page of latest post
	page = pages.dataset.current ? "1"
	date = latest.querySelector('a.date').dataset.udate
	# id of latest post
	if date?
		surl = stripBasepath(window.location.pathname).split('/')[1]
		window.localStorage.setItem "lview_#{surl}", "#{date}##{npost}##{page}"
else if pageIs.home or pageIs.threads
	# Bind "Mark All Read" button
	document.getElementById('mark-all-read').addEventListener 'click', markAllRead

	# In home/all-threads: for each thread, check if already viewed or not
	window.fromList(document.querySelectorAll 'article.thread-item').map (thread) ->
		# get last-modified date
		date = thread.querySelector('span.date').dataset.udate
		# compare with locally saved date, if any
		lreplyAnchor = thread.querySelector 'a.last-reply'
		splpath = stripBasepath(lreplyAnchor.pathname).split('/')
		surl = splpath[1]
		item = window.localStorage.getItem "lview_#{surl}"
		return unless item?
		[lview, lpost, lpage] = item.split '#'
		if lview? and lview == date
			# no updates since latest visit
			if SiteOptions.dehighlight
				thread.className += " seen"
		else if lpost?
			# make the last reply link to point to the latest seen post
			if SiteOptions.jumptolastread
				if splpath.length > 3 and lpage?
					# pathname contains /page/n
					lreplyAnchor.pathname = splpath[0..1].join('/') + "/page/#{lpage}"
				lreplyAnchor.hash = if lpost is 0 then "#thread" else "#p#{lpost}"
