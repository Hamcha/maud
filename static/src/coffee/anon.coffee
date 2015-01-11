anons = ["Rock", "Boulder", "Pebble", "Mineral", "Stone"]

randomAnon = () -> anons[Math.floor(Math.random()*anons.length)]

fromList(document.querySelectorAll ".thread-author,.post-author").map (elem) ->
	nick = elem.querySelector ".nickname"
	trip = elem.querySelector ".tripcode"
	return if nick.innerHTML.length + trip.innerHTML.length > 0
	nick.innerHTML = "<span class=\"anon\">"+randomAnon()+"</span>"
