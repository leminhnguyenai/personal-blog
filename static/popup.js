const popup = {
    process: null,
}

function process(body) {
    const popups = body.querySelectorAll('.pop-up')
    const main = body.querySelector('#main')

    const mainRect = main.getBoundingClientRect()

    popups.forEach((popup) => {
        const handler = () => {
            const viewportHeight = window.innerHeight
            const rect = popup.parentNode.getBoundingClientRect()

            const bottom = viewportHeight - rect.bottom - popup.offsetHeight

            if (bottom < 10) {
                popup.classList.replace('pop-up-bottom', 'pop-up-top')
            } else {
                popup.classList.replace('pop-up-top', 'pop-up-bottom')
            }
        }

        body.removeEventListener('resize', handler)
        main.removeEventListener('scroll', handler)
        body.addEventListener('resize', handler)
        main.addEventListener('scroll', handler)
    })
}

popup.process = process
