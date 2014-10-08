pad = (n) -> ("0"+n).slice(-2)

format = (date) ->
    fulldate = [date.getDate(), date.getMonth()+1, date.getFullYear()].join("/")
    fullhour = [date.getHours(), date.getMinutes()].map(pad).join(":")
    return fulldate + " " + fullhour

sincetime = (diff) ->
    int = Math.floor(diff/86400)
    return int + " days ago" if int > 1
    int = Math.floor(diff/3600)
    return int + " hours ago" if int > 1
    int = Math.floor(diff/60)
    return int + " minutes ago" if int > 1
    return int + " seconds ago"

since = (time) ->
    now = (new Date).getTime()
    diff = now-time

    switch
        when diff > 604800 then format(new Date(time))
        when diff > 3600   then sincetime(diff)+" ("+format(new Date(time))+")"
        else                    sincetime(diff)

elems = document.querySelectorAll ".date"
for elem in elems
    elem.innerHTML = since(parseInt(elem.innerHTML)*1000)