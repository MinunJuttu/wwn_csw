document.addEventListener("DOMContentLoaded", () => {
    setupDynamicList({
        listID: "foci-list",
        templateID: "focus-template",
        buttonID: "add-focus",
        maxItems: 10,
        maxText: "Максимум 10 черт",
    });

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
        maxItems: 5,
        maxText: "Максимум 5 традиций",
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
    const addButton = document.getElementById(buttonID);

    if (!list || !template || !addButton) {
        return;
    }

    const normalButtonText =
        addButton.textContent.trim();


    function getEntries() {
        return Array.from(
            list.querySelectorAll(":scope > .dynamic-entry")
        );
    }


    function reindexEntries() {
        const entries = getEntries();

        entries.forEach((entry, index) => {
            const fields =
                entry.querySelectorAll("[name]");

            fields.forEach((field) => {
                field.name = field.name.replace(
                    /_\d+$/,
                    `_${index}`
                );
            });
        });

        updateControls();
    }


    function updateControls() {
        const entries = getEntries();
        const count = entries.length;

        addButton.disabled = count >= maxItems;

        addButton.textContent =
            count >= maxItems
                ? maxText
                : normalButtonText;


        entries.forEach((entry) => {
            const removeButton =
                entry.querySelector(
                    ".remove-entry-button"
                );

            if (!removeButton) {
                return;
            }

            removeButton.disabled =
                count <= 1;
        });
    }


    addButton.addEventListener("click", () => {
        const entries = getEntries();

        if (entries.length >= maxItems) {
            return;
        }

        const index = entries.length;

        const html =
            template.innerHTML.replaceAll(
                "__INDEX__",
                String(index)
            );

        list.insertAdjacentHTML(
            "beforeend",
            html
        );

        reindexEntries();

        const newEntry =
            getEntries().at(-1);

        if (newEntry) {
            const firstField =
                newEntry.querySelector(
                    "input, textarea"
                );

            if (firstField) {
                firstField.focus();
            }
        }
    });


    list.addEventListener("click", (event) => {
        const removeButton =
            event.target.closest(
                ".remove-entry-button"
            );

        if (!removeButton) {
            return;
        }

        const entries = getEntries();

        if (entries.length <= 1) {
            return;
        }

        const entry =
            removeButton.closest(
                ".dynamic-entry"
            );

        if (!entry) {
            return;
        }

        entry.remove();

        reindexEntries();
    });


    /*
        Это особенно важно для старых
        сохранённых персонажей.

        Если индексы когда-либо были:
        0, 1, 4, 7

        после загрузки получаем:
        0, 1, 2, 3
    */

    reindexEntries();
}