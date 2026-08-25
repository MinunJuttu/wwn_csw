document.addEventListener("DOMContentLoaded", () => {
    setupDynamicList({
        listID: "spells-list",
        templateID: "spell-template",
        buttonID: "add-spell",
        maxItems: 50,
        maxText: "Максимум 50 заклинаний",
    });

    setupDynamicList({
        listID: "magic-arts-list",
        templateID: "magic-art-template",
        buttonID: "add-magic-art",
        maxItems: 25,
        maxText: "Максимум 25 искусств",
    });

    setupDynamicList({
        listID: "magic-traditions-list",
        templateID: "magic-tradition-template",
        buttonID: "add-magic-tradition",
        maxItems: 50,
        maxText: "Достигнут максимум традиций",
    });
});


function setupDynamicList({
    listID,
    templateID,
    buttonID,
    maxItems,
    maxText,
}) {
    const list = document.getElementById(listID);
    const template = document.getElementById(templateID);
    const button = document.getElementById(buttonID);

    if (!list || !template || !button) {
        return;
    }

    let nextIndex = Number(
        list.dataset.nextIndex || 0
    );

    updateButton();

    button.addEventListener("click", () => {
        if (nextIndex >= maxItems) {
            return;
        }

        const html = template.innerHTML.replaceAll(
            "__INDEX__",
            String(nextIndex)
        );

        list.insertAdjacentHTML(
            "beforeend",
            html
        );

        nextIndex++;

        updateButton();

        const lastEntry = list.lastElementChild;

        if (lastEntry) {
            const firstInput =
                lastEntry.querySelector(
                    "input, textarea"
                );

            if (firstInput) {
                firstInput.focus();
            }
        }
    });


    function updateButton() {
        if (nextIndex >= maxItems) {
            button.disabled = true;
            button.textContent = maxText;
        }
    }
}