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

    const weaponList = setupDynamicList({
        listID: "weapons-list",
        templateID: "weapon-template",
        buttonID: "add-weapon",
        maxItems: 20,
        maxText: "Максимум 20 единиц оружия",
    });

    setupWeaponPresets(weaponList);
    setupWeaponHelp();
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
        return null;
    }

    const normalButtonText = addButton.textContent.trim();

    function getEntries() {
        return Array.from(list.children).filter((element) =>
            element.classList.contains("dynamic-entry")
        );
    }

    function reindexEntries() {
        const entries = getEntries();

        entries.forEach((entry, index) => {
            const fields = entry.querySelectorAll("[name]");

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
            const removeButton = entry.querySelector(
                ".remove-entry-button"
            );

            if (!removeButton) {
                return;
            }

            removeButton.disabled = count <= 1;
        });
    }

    function addEntry() {
        const entries = getEntries();

        if (entries.length >= maxItems) {
            return null;
        }

        const index = entries.length;
        const html = template.innerHTML.replaceAll(
            "__INDEX__",
            String(index)
        );

        list.insertAdjacentHTML("beforeend", html);
        reindexEntries();

        return getEntries().at(-1) || null;
    }

    addButton.addEventListener("click", () => {
        const newEntry = addEntry();

        if (!newEntry) {
            return;
        }

        const firstField = newEntry.querySelector(
            "input, textarea"
        );

        if (firstField) {
            firstField.focus();
        }
    });

    list.addEventListener("click", (event) => {
        const removeButton = event.target.closest(
            ".remove-entry-button"
        );

        if (!removeButton) {
            return;
        }

        const entries = getEntries();

        if (entries.length <= 1) {
            return;
        }

        const entry = removeButton.closest(
            ".dynamic-entry"
        );

        if (!entry) {
            return;
        }

        entry.remove();
        reindexEntries();
    });

    reindexEntries();

    return {
        addEntry,
        getEntries,
        reindexEntries,
        maxItems,
    };
}


function setupWeaponPresets(weaponList) {
    const picker = document.getElementById(
        "weapon-preset-picker"
    );
    const menu = document.getElementById(
        "weapon-preset-menu"
    );

    if (!picker || !menu || !weaponList) {
        return;
    }

    const reference = window.WWN_WEAPON_REFERENCE;
    const presets = Array.isArray(reference?.presets)
        ? reference.presets
        : [];

    menu.replaceChildren();

    if (presets.length === 0) {
        const empty = document.createElement("div");
        empty.className = "weapon-preset-empty";
        empty.textContent = "Список стандартного оружия пока пуст.";
        menu.appendChild(empty);
        return;
    }

    presets.forEach((preset) => {
        const button = document.createElement("button");
        button.type = "button";
        button.className = "weapon-preset-option";

        const name = document.createElement("strong");
        name.textContent = preset.name || "Без названия";
        button.appendChild(name);

        const details = weaponPresetDetails(preset);
        if (details) {
            const meta = document.createElement("span");
            meta.textContent = details;
            button.appendChild(meta);
        }

        button.addEventListener("click", () => {
            const emptyEntry = weaponList
                .getEntries()
                .find(isWeaponEntryEmpty);

            const entry = emptyEntry || weaponList.addEntry();

            if (!entry) {
                picker.open = false;
                return;
            }

            setWeaponField(entry, "weapon_name_", preset.name);
            setWeaponField(entry, "weapon_attribute_", preset.attribute);
            setWeaponField(
                entry,
                "weapon_encumbrance_",
                preset.encumbrance
            );
            setWeaponField(entry, "weapon_hit_", preset.hitBonus);
            setWeaponField(entry, "weapon_damage_", preset.damage);
            setWeaponField(entry, "weapon_range_", preset.range);
            setWeaponField(
                entry,
                "weapon_special_",
                preset.specialShock
            );

            picker.open = false;

            const nameInput = entry.querySelector(
                '[name^="weapon_name_"]'
            );

            if (nameInput) {
                nameInput.focus();
            }
        });

        menu.appendChild(button);
    });
}


function isWeaponEntryEmpty(entry) {
    const fields = entry.querySelectorAll(
        'input[name^="weapon_"]'
    );

    return Array.from(fields).every(
        (field) => field.value.trim() === ""
    );
}


function setWeaponField(entry, prefix, value) {
    const field = entry.querySelector(
        `[name^="${prefix}"]`
    );

    if (!field) {
        return;
    }

    field.value = value ?? "";
}


function weaponPresetDetails(preset) {
    const parts = [];

    if (preset.attribute) {
        parts.push(preset.attribute);
    }

    if (preset.encumbrance) {
        parts.push(`нагрузка ${preset.encumbrance}`);
    }

    if (preset.damage) {
        parts.push(`урон ${preset.damage}`);
    }

    if (preset.range) {
        parts.push(`дальность ${preset.range}`);
    }

    return parts.join(" · ");
}


function setupWeaponHelp() {
    const openButton = document.getElementById(
        "weapon-help-button"
    );
    const dialog = document.getElementById(
        "weapon-help-dialog"
    );
    const content = document.getElementById(
        "weapon-help-content"
    );

    if (!openButton || !dialog || !content) {
        return;
    }

    renderWeaponHelp(content);

    openButton.addEventListener("click", () => {
        if (typeof dialog.showModal === "function") {
            dialog.showModal();
            return;
        }

        dialog.setAttribute("open", "");
    });

    dialog.querySelectorAll("[data-close-dialog]").forEach(
        (button) => {
            button.addEventListener("click", () => {
                dialog.close();
            });
        }
    );

    dialog.addEventListener("click", (event) => {
        if (event.target === dialog) {
            dialog.close();
        }
    });
}


function renderWeaponHelp(container) {
    const help = window.WWN_WEAPON_REFERENCE?.help || {};

    container.replaceChildren();

    container.appendChild(
        createHelpSection(
            "Бросок атаки",
            help.attackRoll || "Справка пока не заполнена."
        )
    );

    container.appendChild(
        createHelpSection(
            "Бросок урона",
            help.damageRoll || "Справка пока не заполнена."
        )
    );

    const tagsSection = document.createElement("section");
    tagsSection.className = "weapon-help-section";

    const heading = document.createElement("h3");
    heading.textContent = "Теги оружия";
    tagsSection.appendChild(heading);

    const tags = Array.isArray(help.tags) ? help.tags : [];

    if (tags.length === 0) {
        const empty = document.createElement("p");
        empty.className = "weapon-help-text";
        empty.textContent = "Справка по тегам пока не заполнена.";
        tagsSection.appendChild(empty);
    } else {
        const list = document.createElement("div");
        list.className = "weapon-help-tags";

        tags.forEach((tag) => {
            const item = document.createElement("div");
            item.className = "weapon-help-tag";

            const name = document.createElement("strong");
            name.textContent = tag.name || "Тег";

            const description = document.createElement("p");
            description.className = "weapon-help-text";
            description.textContent = tag.description || "";

            item.append(name, description);
            list.appendChild(item);
        });

        tagsSection.appendChild(list);
    }

    container.appendChild(tagsSection);
}


function createHelpSection(title, text) {
    const section = document.createElement("section");
    section.className = "weapon-help-section";

    const heading = document.createElement("h3");
    heading.textContent = title;

    const paragraph = document.createElement("p");
    paragraph.className = "weapon-help-text";
    paragraph.textContent = text;

    section.append(heading, paragraph);
    return section;
}
