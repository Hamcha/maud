pad = (n) -> ("0"+n).slice(-2)
plur = (x, n) -> if n > 1 then x+"s" else x

format = (date) ->
	fulldate = [date.getDate(), date.getMonth()+1, date.getFullYear()].join("/")
	fullhour = [date.getHours(), date.getMinutes()].map(pad).join(":")
	return fulldate + " " + fullhour

sincetime = (diff) ->
	return "just now" if diff < 1
	int = Math.floor(diff/86400)
	return int + " "+plur("day",int)+" ago" if int > 0
	int = Math.floor(diff/3600)
	return int + " "+plur("hour",int)+" ago" if int > 0
	int = Math.floor(diff/60)
	return int + " "+plur("minute",int)+" ago" if int > 0
	return Math.floor(diff) + " seconds ago"

since = (time, brief) ->
	now = (new Date).getTime()
	diff = (now-time)/1000

	switch
		when diff > 604800          then format(new Date(time))
		when diff > 3600 && !brief  then sincetime(diff)+" ("+format(new Date(time))+")"
		else                        sincetime(diff)

fromList(document.querySelectorAll ".date").map (elem) ->
	elem.innerHTML = since parseInt(elem.dataset.udate)*1000, false

fromList(document.querySelectorAll ".lastedit").map (elem) ->
	if elem.innerHTML is "0"
		elem.style.visibility = 'hidden'
		return
	elem.innerHTML = "edited #{since(parseInt(elem.dataset.udate)*1000, true)}"
