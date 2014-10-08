anons = ["Rock", "Boulder", "Pebble", "Mineral", "Stone"]

randomAnon = () -> anons[Math.floor(Math.random()*anons.length)]

authors = document.querySelectorAll ".thread-author,.post-author"
for elem in authors
    nick = elem.querySelector ".nickname"
    trip = elem.querySelector ".tripcode"
    continue if nick.innerHTML.length + trip.innerHTML.length > 0
    nick.innerHTML = "<span class=\"anon\">"+randomAnon()+"</span>"