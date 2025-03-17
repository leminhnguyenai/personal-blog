const clipboard = {
    process: null,
}

function codeblockCopy(body) {
    const codeblocks = body.querySelectorAll('.codeblock')

    codeblocks.forEach((codeblock) => {
        const code = codeblock.querySelector('pre').textContent
        const clipboardBtn = codeblock.querySelector('button.clipboard')

        const handler = (ev) => {
            navigator.clipboard.writeText(code)
            clipboardBtn.textContent = '󰄬'
            setTimeout(() => {
                clipboardBtn.textContent = ''
            }, 1000)

            const notiEvent = new CustomEvent('noti', {
                detail: {
                    message: 'Code copied to clipboard',
                    status: 'successful',
                },
            })

            clipboardBtn.dispatchEvent(notiEvent)
        }

        clipboardBtn.removeEventListener('click', handler)
        clipboardBtn.addEventListener('click', handler)
    })
}

function headingCopy(body) {
    const main = body.querySelector('#main')
    const content = main.querySelector('#content')
    const headings = content.querySelectorAll('h1, h2, h3, h4, h5')
    const baseURL = window.location.toString().split('#')[0]

    headings.forEach((heading) => {
        const url = baseURL + '#' + heading.id

        const handler = () => {
            const notEvent = new CustomEvent('noti', {
                detail: {
                    message: 'Url copied to clipboard',
                    status: 'warning',
                },
            })

            heading.dispatchEvent(notEvent)

            navigator.clipboard.writeText(url)
        }

        heading.removeEventListener('click', handler)
        heading.addEventListener('click', handler)
    })
}

function process(body) {
    codeblockCopy(body)
    headingCopy(body)
}

clipboard.process = process
