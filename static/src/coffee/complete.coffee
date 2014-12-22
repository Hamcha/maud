###
  AJAX tags autocomplete plugin
  (requires qwest)
###

# autocomplete input from a JSON list (retreived via AJAX)
# opts:
#   - minChars
autocomplete = (elem, url, opts) ->
    ELEM = document.querySelector elem
    return unless ELEM?
    # get the JSON from the server
    data = []
    qwest.post(url, null, { responseType: 'json', async: false })
        .then (resp) ->
            data = resp
        .catch (err) ->
            console.log 'Error retreiving data'
    # element holding the autocomplete data
    ul = document.createElement 'ul'
    ul.className = 'autocomplete-list'
    insertAfter = (newNode, node) ->
    	node.parentNode.insertBefore(newNode, node.nextSibling)
    insertAfter ul, ELEM
    ELEM.onkeyup = (e) ->
        #if opts.keyupFunc?
        #    return if opts.keyupFunc(e)
        if !opts.minChars? || ELEM.value.length >= opts.minChars
            updateAutocompleteList ul, ELEM.value, data
        else
            ul.innerHTML = ""

updateAutocompleteList = (list, txt, data) ->
    list.innerHTML = data.map((el) -> if el.startsWith(txt) then "<li>#{el}</li>").sort().join "\n"

autocomplete '#tagsearch', '/taglist', { minChars: 2 }
# TODO: integrate with Taggle
#autocomplete '.taggle_input', '/taglist', {
#    minChars: 2,
#    keyupFunc: (key) ->
#        if key.keyCode is 188 # comma
#            newThreadTaggle.add(document.querySelector('.taggle_input').value)
#            return true
#        return false
#}
