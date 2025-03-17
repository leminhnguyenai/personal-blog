const notification = {
    process: null,
}

function process(body) {
    const notification = body.querySelector('#notification')
    const senders = body.querySelectorAll('[noti="true"]')

    senders.forEach((sender) => {
        const handler = (ev) => {
            const message = ev.detail.message

            const template = notification.querySelector('template').content.querySelector('p')
            const newPopup = template.cloneNode(true)
            console.log(newPopup)

            newPopup.textContent = message

            notification.insertBefore(newPopup, notification.firstChild)

            setTimeout(() => newPopup.remove(), 3000)
        }

        sender.removeEventListener('noti', handler)
        sender.addEventListener('noti', handler)
    })
}

notification.process = process
