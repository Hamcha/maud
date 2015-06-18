# Allows editing blacklist rule 'name' in HTML element #bl_name
original = {}

blacklistEdit = (name) ->
	elem = document.getElementById "bl_#{name}"
	return unless elem?
	original[name] = elem.innerHTML
	qwest.post(window.stripPage(location.pathname) + "/" + name + "/raw")
		.then (resp) ->
			textarea = elem.querySelector "textarea[name='json']"
			textarea.value = resp
			textarea.placeholder = "Insert ban rule as JSON"
		.catch (err) ->
			content = ""
			return if elem.firstChild.className == "errmsg"
			errmsg = document.createElement 'p'
			errmsg.className = "errmsg"
			errmsg.innerHTML = "Failed to retrieve content: #{err}"
			elem.insertBefore errmsg, elem.firstChild

	elem.innerHTML = """
<form method="POST" action="#{window.stripPage(location.pathname) + "/" + name + "/edit"}">
    <div>
        <span style="color: #ccc; display: inline-block; width: auto; font-size: 0.9em;">editing rule #{name}</span>
    </div>
    <textarea class="full small editor" name="json" required placeholder="Retreiving content..."></textarea>
    <div class="center">
        <input type="Submit" value="Edit rule"/><button type="button" onclick="Blacklist.cancelForm('#{name}');">Cancel</button>
    </div>
</form>
"""

blacklistUnban = (name) ->
	elem = document.getElementById "bl_#{name}"
	return unless elem?
	original[name] = elem.innerHTML
	elem.innerHTML = """
<form method="POST" action="#{window.stripPage(location.pathname) + "/" + name + "/delete"}">
    <div>
        <span style="color: #ccc; display: inline-block; width: auto; font-size: 0.9em;">deleting rule #{name}</span>
    </div>
    <div class="center">
        <button name="deletetype" value="soft" type="submit">Unban</button><button type="button" onclick="Blacklist.cancelForm('#{name}');">Cancel</button>
    </div>
</form>
"""

cancelForm = (name) ->
	elem = document.getElementById "bl_#{name}"
	elem.innerHTML = original[name]


window.Blacklist =
	edit: blacklistEdit
	unban: blacklistUnban
	cancelForm: cancelForm
