// Add clipboard functionality to the clipboard button
const codeblocks = document.querySelectorAll('.codeblock')

codeblocks.forEach((codeblock) => {
    const code = codeblock.querySelector('pre').textContent
    const clipboardBtn = codeblock.querySelector('button.clipboard')

    clipboardBtn.addEventListener('click', () => {
        navigator.clipboard.writeText(code)
        clipboardBtn.textContent = '󰄬'
        setTimeout(() => {
            clipboardBtn.textContent = ''
        }, 1000)
    })
})

// Highlight TOC when section is visible
const main = document.querySelector('#main')
const headings = main.querySelectorAll('h1, h2, h3, h4, h5')
const toc = document.querySelector('.side-bar.toc')
const chapters = toc.querySelectorAll('a.chapter')

const sections = []

function removeClass(el, className) {
    if (el.classList.contains(className)) {
        el.classList.remove(className)
    }
}

function addClass(el, className) {
    if (!el.classList.contains(className)) {
        el.classList.add(className)
    }
}

for (let i = 0; i < headings.length; i++) {
    sections.push({
        heading: headings[i],
        chapter: chapters[i],
    })
}

main.addEventListener('scroll', () => {
    for (let i = 0; i < sections.length; i++) {
        if (i == sections.length - 1) {
            const rect = sections[i].heading.getBoundingClientRect()
            if (rect.top <= 0) {
                addClass(sections[i].chapter, 'chapter-highlight')

                for (let j = 0; j < sections.length; j++) {
                    if (j != i) {
                        removeClass(sections[j].chapter, 'chapter-highlight')
                    }
                }
                break
            }
            continue
        }

        const upperRect = sections[i].heading.getBoundingClientRect()
        const lowerRect = sections[i + 1].heading.getBoundingClientRect()

        if (upperRect.top <= 0 && lowerRect.top > 0) {
            addClass(sections[i].chapter, 'chapter-highlight')

            for (let j = 0; j < sections.length; j++) {
                if (j != i) {
                    removeClass(sections[j].chapter, 'chapter-highlight')
                }
            }
            break
        }
    }
})
