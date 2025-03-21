function getDocument() {
    return document
}

// Go through all nodes and apply appropriate handlers
function processNodes(body) {
    // Import all modules here
    notification.process(body)
    clipboard.process(body)
    code.process(body)
    popup.process(body)
}

//= ===================================================================
// Initialization
//= ===================================================================
var isReady = false
getDocument().addEventListener('DOMContentLoaded', function () {
    isReady = true
})

function ready(fn) {
    if (isReady || getDocument().readyState === 'complete') {
        fn()
    } else {
        getDocument().addEventListener('DOMContentLoaded', fn)
    }
}

ready(() => {
    let body = getDocument().body
    processNodes(body)
})
