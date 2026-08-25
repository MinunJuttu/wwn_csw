document.addEventListener("DOMContentLoaded", () => {
    const picker = document.getElementById("character-picker");
    const closePicker = document.getElementById("close-character-picker");
    const filter = document.getElementById("character-picker-filter");
    const items = Array.from(document.querySelectorAll(".character-picker-item"));
    const slots = Array.from(document.querySelectorAll(".gm-slot"));

    if (!picker || slots.length === 0) {
        return;
    }

    let activeSlot = null;

    slots.forEach((slot) => {
        const emptyButton = slot.querySelector(".gm-slot-empty");
        const removeButton = slot.querySelector(".gm-slot-remove");

        emptyButton?.addEventListener("click", () => {
            activeSlot = slot;
            filter.value = "";
            applyFilter("");
            picker.showModal();
            setTimeout(() => filter.focus(), 0);
        });

        removeButton?.addEventListener("click", () => {
            clearSlot(slot);
            saveSlots();
        });
    });

    items.forEach((item) => {
        item.addEventListener("click", () => {
            if (!activeSlot) {
                return;
            }

            fillSlot(activeSlot, {
                id: item.dataset.characterId,
                name: item.dataset.characterName,
                owner: item.dataset.characterOwner,
                level: item.dataset.characterLevel,
            });

            saveSlots();
            picker.close();
            activeSlot = null;
        });
    });

    closePicker?.addEventListener("click", () => {
        picker.close();
        activeSlot = null;
    });

    picker.addEventListener("click", (event) => {
        if (event.target === picker) {
            picker.close();
            activeSlot = null;
        }
    });

    filter?.addEventListener("input", () => {
        applyFilter(filter.value);
    });

    restoreSlots();

    function applyFilter(value) {
        const needle = value.trim().toLowerCase();

        items.forEach((item) => {
            const haystack = `${item.dataset.characterName} ${item.dataset.characterOwner}`.toLowerCase();
            item.hidden = needle !== "" && !haystack.includes(needle);
        });
    }

    function fillSlot(slot, character) {
        const empty = slot.querySelector(".gm-slot-empty");
        const filled = slot.querySelector(".gm-slot-filled");
        const title = slot.querySelector(".gm-slot-title");
        const frame = slot.querySelector(".gm-slot-frame");

        slot.dataset.characterId = character.id;
        slot.dataset.characterName = character.name;
        slot.dataset.characterOwner = character.owner;
        slot.dataset.characterLevel = character.level;

        title.innerHTML = `
            <strong>${escapeHTML(character.name)}</strong>
            <span>${escapeHTML(character.owner)} · уровень ${escapeHTML(character.level)}</span>
        `;

        frame.src = `/admin/characters/${encodeURIComponent(character.id)}?embed=1`;
        empty.hidden = true;
        filled.hidden = false;
    }

    function clearSlot(slot) {
        const empty = slot.querySelector(".gm-slot-empty");
        const filled = slot.querySelector(".gm-slot-filled");
        const title = slot.querySelector(".gm-slot-title");
        const frame = slot.querySelector(".gm-slot-frame");

        delete slot.dataset.characterId;
        delete slot.dataset.characterName;
        delete slot.dataset.characterOwner;
        delete slot.dataset.characterLevel;

        title.textContent = "";
        frame.removeAttribute("src");
        filled.hidden = true;
        empty.hidden = false;
    }

    function saveSlots() {
        const state = slots.map((slot) => {
            if (!slot.dataset.characterId) {
                return null;
            }

            return {
                id: slot.dataset.characterId,
                name: slot.dataset.characterName,
                owner: slot.dataset.characterOwner,
                level: slot.dataset.characterLevel,
            };
        });

        localStorage.setItem("wwn-admin-table", JSON.stringify(state));
    }

    function restoreSlots() {
        let state;

        try {
            state = JSON.parse(localStorage.getItem("wwn-admin-table") || "[]");
        } catch {
            state = [];
        }

        if (!Array.isArray(state)) {
            return;
        }

        state.slice(0, slots.length).forEach((character, index) => {
            if (!character || !character.id) {
                return;
            }

            const stillExists = items.some(
                (item) => item.dataset.characterId === String(character.id)
            );

            if (stillExists) {
                fillSlot(slots[index], character);
            }
        });
    }

    function escapeHTML(value) {
        return String(value ?? "")
            .replaceAll("&", "&amp;")
            .replaceAll("<", "&lt;")
            .replaceAll(">", "&gt;")
            .replaceAll('"', "&quot;")
            .replaceAll("'", "&#039;");
    }
});
