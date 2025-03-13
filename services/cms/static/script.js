const codeblocks = document.querySelectorAll('.codeblock')

// Add clipboard functionality to the clipboard button
codeblocks.forEach((codeblock) => {
    const code = codeblock.querySelector('pre').textContent
    const clipboardBtn = codeblock.querySelector('button.clipboard')

    clipboardBtn.addEventListener('click', () => {
        navigator.clipboard.writeText(code)
        clipboardBtn.textContent = '󰄬'
        setTimeout(() => {
            clipboardBtn.textContent = ''
        }, 3000)
    })
})
