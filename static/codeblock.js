const code = {
    process: null,
}

function process(body) {
    const codeblocks = body.querySelectorAll('.codeblock')

    codeblocks.forEach((codeblock) => {
        const gutter = codeblock.querySelectorAll('p.code-gutter')
        const code = codeblock.querySelectorAll('p.code-line')

        const resizeObserver = new ResizeObserver(() => {
            for (let i = 0; i < code.length; i++) {
                if (gutter[i].offsetHeight != code[i].offsetHeight) {
                    gutter[i].style.height = `${code[i].offsetHeight}px`
                }
            }
        })

        resizeObserver.observe(codeblock)
    })
}

code.process = process
