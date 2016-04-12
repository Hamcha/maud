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

mergeProps = (a, b) ->
	for key, val of b
		# Style is picky, can't recurse
		if key is "style"
			a.style[prop] = value for prop, value of val
		else if typeof val is "object" and a[key]?
			a[key] = mergeProps a[key], val
		else
			a[key] = val
	return a

window.createElementEx = (elemName, elemProps) ->
	element = document.createElement elemName
	if elemProps?
		element = mergeProps element, elemProps
	return element
