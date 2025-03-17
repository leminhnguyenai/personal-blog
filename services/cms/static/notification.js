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
                    template = notification.querySelector('template').content.querySelector('p.successful')
                    break
                case 'warning':
                    template = notification.querySelector('template').content.querySelector('p.warning')
                    break
                case 'error':
                    template = notification.querySelector('template').content.querySelector('p.error')
                    break
                default:
                    template = notification.querySelector('template').content.querySelector('p.default')
                    break
            }

            const newPopup = template.cloneNode(true)

            newPopup.textContent = message

            notification.insertBefore(newPopup, notification.firstChild)

            setTimeout(() => newPopup.remove(), 3000)
        }

        sender.removeEventListener('noti', handler)
        sender.addEventListener('noti', handler)
    })
}

notification.process = process
