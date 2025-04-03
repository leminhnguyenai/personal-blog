// ======== HELPERS ======== /
function removeClass(el, className) {
    if (el.classList.contains(className)) {
        el.classList.remove(className);
    }
}

function addClass(el, className) {
    if (!el.classList.contains(className)) {
        el.classList.add(className);
    }
}

function getDocument() {
    return document;
}

// ======== CLIPBOARD ======== /
const clipboard = {
    codeblockCopy: (body) => {
        const codeblocks = body.querySelectorAll('[codeblock]');

        codeblocks.forEach((codeblock) => {
            const code = codeblock.querySelector('pre').textContent;
            const clipboardBtn = codeblock.querySelector('button[clipboard]');

            const handler = (ev) => {
                navigator.clipboard.writeText(code);
                clipboardBtn.textContent = '󰄬';
                setTimeout(() => {
                    clipboardBtn.textContent = '';
                }, 1000);

                const notiEvent = new CustomEvent('noti', {
                    detail: {
                        message: 'Code copied to clipboard',
                        status: 'successful',
                    },
                });

                clipboardBtn.dispatchEvent(notiEvent);
            };

            clipboardBtn.removeEventListener('click', handler);
            clipboardBtn.addEventListener('click', handler);
        });
    },

    headingCopy: (body) => {
        const content = body.querySelector('#content');
        const headings = content.querySelectorAll('h1, h2, h3, h4, h5');
        const baseURL = window.location.toString().split('#')[0];

        headings.forEach((heading) => {
            const url = baseURL + '#' + heading.id;

            const handler = () => {
                const notEvent = new CustomEvent('noti', {
                    detail: {
                        message: 'Url copied to clipboard',
                        status: 'successful',
                    },
                });

                heading.dispatchEvent(notEvent);

                navigator.clipboard.writeText(url);
            };

            heading.removeEventListener('click', handler);
            heading.addEventListener('click', handler);
        });
    },
    process: function (body) {
        this.codeblockCopy(body);
        this.headingCopy(body);
    },
};

// ======== CODEBLOCK ======== /
const code = {
    process: function (body) {
        const codeblocks = body.querySelectorAll('[codeblock]');

        codeblocks.forEach((codeblock) => {
            const gutter = codeblock.querySelectorAll('p[code-gutter]');
            const code = codeblock.querySelectorAll('p[code-line]');

            const resizeObserver = new ResizeObserver(() => {
                for (let i = 0; i < code.length; i++) {
                    if (gutter[i].offsetHeight != code[i].offsetHeight) {
                        gutter[i].style.height = `${code[i].offsetHeight}px`;
                    }
                }
            });

            resizeObserver.observe(codeblock);
        });
    },
};

// ======== NOTIFICATION ======== /
const notification = {
    process: function (body) {
        const notification = body.querySelector('#notification');
        const senders = body.querySelectorAll('[noti="true"]');

        senders.forEach((sender) => {
            const handler = (ev) => {
                const message = ev.detail.message;
                const status = ev.detail.status;
                var template;

                switch (status) {
                    case 'successful':
                        template = notification.querySelector('template').content.querySelector('.successful');
                        break;
                    case 'warning':
                        template = notification.querySelector('template').content.querySelector('.warning');
                        break;
                    case 'error':
                        template = notification.querySelector('template').content.querySelector('.error');
                        break;
                    default:
                        template = notification.querySelector('template').content.querySelector('.default');
                        break;
                }

                const newPopup = template.cloneNode(true);

                newPopup.querySelector('p').textContent = message;
                newPopup.querySelector('span').addEventListener('click', () => newPopup.remove());

                notification.insertBefore(newPopup, notification.firstChild);

                setTimeout(() => newPopup.remove(), 5000);
            };

            sender.removeEventListener('noti', handler);
            sender.addEventListener('noti', handler);
        });
    },
};

// ======== POPUP ======== /
const popup = {
    process: function (body) {
        const popups = body.querySelectorAll('[pop-up]');

        popups.forEach((popup) => {
            const handler = () => {
                const viewportHeight = window.innerHeight;
                const rect = popup.parentNode.getBoundingClientRect();

                const bottom = viewportHeight - rect.bottom - popup.offsetHeight;

                if (bottom < 10) {
                    popup.classList.replace('pop-up-bottom', 'pop-up-top');
                } else {
                    popup.classList.replace('pop-up-top', 'pop-up-bottom');
                }
            };

            body.removeEventListener('resize', handler);
            body.removeEventListener('scroll', handler);
            body.addEventListener('resize', handler);
            body.addEventListener('scroll', handler);
        });
    },
};

//= ===================================================================
// Initialization
//= ===================================================================
// Go through all nodes and apply appropriate handlers
function processNodes(body) {
    // Import all modules here
    notification.process(body);
    clipboard.process(body);
    code.process(body);
    popup.process(body);
}

var isReady = false;
getDocument().addEventListener('DOMContentLoaded', function () {
    isReady = true;
});

function ready(fn) {
    if (isReady || getDocument().readyState === 'complete') {
        fn();
    } else {
        getDocument().addEventListener('DOMContentLoaded', fn);
    }
}

ready(() => {
    let body = getDocument().body;
    processNodes(body);
});
