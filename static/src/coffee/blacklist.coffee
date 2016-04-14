'use strict'

cancelForm = (elem, original) ->
	return ->
		elem.removeChild elem.firstChild
		elem.appendChild e for e in original

blacklistEdit = (name) ->
	return ->
		elem = document.getElementById "bl_#{name}"
		return unless elem?
		original = Array.prototype.slice.apply elem.children
		console.log original

		action = (window.stripPage location.pathname) + "/" + name + "/edit"
		form = window.createElementEx "form", { method: "POST", action: action }

		div = window.createElementEx "div"
		form.appendChild div

		span = window.createElementEx "span", { className: "bl-edit" }
		span.appendChild document.createTextNode "editing rule #{name}"
		div.appendChild span

		textarea = window.createElementEx "textarea", {
			className: "full small editor",
			name: "json",
			required: true,
			placeholder: "Retrieving content..."
		}
		form.appendChild textarea

		centered = window.createElementEx "div", { className: "center" }
		form.appendChild centered

		submit = window.createElementEx "input", { type: "submit", value: "Edit rule" }
		centered.appendChild submit

		cancel = window.createElementEx "button", { type: "button" }
		cancel.addEventListener "click", cancelForm elem, original
		cancel.appendChild document.createTextNode "Cancel"
		centered.appendChild cancel

		elem.removeChild elem.firstChild while elem.firstChild?
		elem.appendChild form

		qwest.post(window.stripPage(location.pathname) + "/" + name + "/raw")
			.then (resp) ->
				textarea.value = resp
				textarea.placeholder = "Insert ban rule as JSON"
			.catch (err) ->
				content = ""
				return if elem.firstChild.className == "errmsg"
				errmsg = window.createElementEx 'p', { className: "errmsg" }
				errmsg.appendChild document.createTextNode "Failed to retrieve content: #{err}"
				elem.insertBefore errmsg, elem.firstChild

blacklistUnban = (name) ->
	return ->
		elem = document.getElementById "bl_#{name}"
		return unless elem?
		original = Array.prototype.slice.apply elem.children

		action = (window.stripPage location.pathname) + "/" + name + "/delete"
		form = window.createElementEx "form", { method: "POST", action: action }

		div = window.createElementEx "div"
		form.appendChild div

		span = window.createElementEx "span", { className: "bl-edit" }
		span.appendChild document.createTextNode "deleting rule #{name}"
		div.appendChild span

		centered = window.createElementEx "div", { className: "center" }
		form.appendChild centered

		submit = window.createElementEx "button", { type: "submit", name: "deletetype", value: "soft" }
		submit.appendChild document.createTextNode "Unban"
		centered.appendChild submit

		cancel = window.createElementEx "button", { type: "button" }
		cancel.addEventListener "click", cancelForm elem, original
		cancel.appendChild document.createTextNode "Cancel"
		centered.appendChild cancel

		elem.removeChild elem.firstChild while elem.firstChild?
		elem.appendChild form

# Bind events
(Array.prototype.slice.apply document.getElementsByClassName 'blacklist-edit-btn').map (e) ->
	e.addEventListener 'click', blacklistEdit e.dataset.name

(Array.prototype.slice.apply document.getElementsByClassName 'blacklist-unban-btn').map (e) ->
	e.addEventListener 'click', blacklistUnban e.dataset.name