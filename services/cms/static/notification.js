const notification = {
    process: null,
}

// COMMIT: Add support for manually closing notification popup
// COMMIT: Add more customization options for notification popup
function process(body) {
    const notification = body.querySelector('#notification')
    const senders = body.querySelectorAll('[noti="true"]')

    senders.forEach((sender) => {
        const handler = (ev) => {
            const message = ev.detail.message
            const status = ev.detail.status
            var template

            switch (status) {
                case 'successful':
                    template = notification.querySelector('template').content.querySelector('.successful')
                    break
                case 'warning':
                    template = notification.querySelector('template').content.querySelector('.warning')
                    break
                case 'error':
                    template = notification.querySelector('template').content.querySelector('.error')
                    break
                default:
                    template = notification.querySelector('template').content.querySelector('.default')
                    break
            }

            const newPopup = template.cloneNode(true)

            newPopup.querySelector('p').textContent = message
            newPopup.querySelector('span').addEventListener('click', () => newPopup.remove())

            notification.insertBefore(newPopup, notification.firstChild)

            setTimeout(() => newPopup.remove(), 5000)
        }

        sender.removeEventListener('noti', handler)
        sender.addEventListener('noti', handler)
    })
}

notification.process = process
