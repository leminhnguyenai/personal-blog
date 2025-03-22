// Highlight TOC when section is visible
const main = document.querySelector('#main');
const headings = main.querySelectorAll('h1, h2, h3, h4, h5');
const toc = document.querySelector('.side-bar.toc');
const chapters = toc.querySelectorAll('a.chapter');
const topbar = document.querySelector('div.top-bar');

const sections = [];

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

for (let i = 0; i < headings.length; i++) {
    sections.push({
        heading: headings[i],
        chapter: chapters[i],
    });
}

const handler = () => {
    for (let i = 0; i < sections.length; i++) {
        if (i == sections.length - 1) {
            const rect = sections[i].heading.getBoundingClientRect();
            if (rect.top <= 0) {
                addClass(sections[i].chapter, 'chapter-highlight');

                for (let j = 0; j < sections.length; j++) {
                    if (j != i) {
                        removeClass(sections[j].chapter, 'chapter-highlight');
                    }
                }
                break;
            }
            continue;
        }

        const upperRect = sections[i].heading.getBoundingClientRect();
        const lowerRect = sections[i + 1].heading.getBoundingClientRect();

        if (upperRect.top <= topbar.offsetHeight && lowerRect.top > topbar.offsetHeight) {
            addClass(sections[i].chapter, 'chapter-highlight');

            sections.forEach((section, index) => {
                if (index != i) {
                    removeClass(section.chapter, 'chapter-highlight');
                }
            });
            break;
        }
    }
};

main.removeEventListener('scroll', handler);
main.addEventListener('scroll', handler);
