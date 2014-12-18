window.fromList = (x) -> Array.prototype.slice.call x

window.stripPage = (url) -> 
	idx = url.indexOf "/page/"
	return if idx < 0 then url else url.substring 0, idx