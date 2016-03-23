'use strict'

window.fromList = (x) -> Array.prototype.slice.call x

window.insertAfter = (newNode, node) ->
	node.parentNode.insertBefore(newNode, node.nextSibling)

window.stripPage = (url) ->
	idx = url.indexOf "/page/"
	return if idx < 0 then url else url.substring 0, idx

window.escapeHTML = (str) ->
	return str
		.replace(/&/g, "&amp;")
		.replace(/\//g, "&sol;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#039;")

window.getViewport = ->
	e = window
	a = 'inner'
	unless 'innerWidth' in window
		a = 'client'
		e = document.documentElement || document.body
	return { width: e["#{a}Width"], height: e["#{a}Height"] }
