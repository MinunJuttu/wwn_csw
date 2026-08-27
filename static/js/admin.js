document.addEventListener("DOMContentLoaded", () => {
    const desk = document.getElementById("gm-desk");
    const picker = document.getElementById("character-picker");
    const closePicker = document.getElementById("close-character-picker");
    const filter = document.getElementById("character-picker-filter");
    const addSlotButton = document.getElementById("gm-add-slot");
    const removeSlotButton = document.getElementById("gm-remove-slot");
    const noiseInput = document.getElementById("gm-noise-level");
    const noiseHelp = document.getElementById("gm-noise-help");
    const items = Array.from(document.querySelectorAll(".character-picker-item"));

    if (!desk || !picker) {
        return;
    }

    const STORAGE_KEY = "wwn-admin-table";
    const MIN_SLOTS = 2;
    const MAX_SLOTS = 5;
    const DEFAULT_SLOTS = 2;

    let activeSlot = null;
    const restoredState = readState();
    const initialSlotCount = clamp(restoredState.slotCount, MIN_SLOTS, MAX_SLOTS);

    for (let index = 0; index < initialSlotCount; index += 1) {
        const slot = createSlot(index);
        desk.appendChild(slot);

        const character = restoredState.slots[index];
        if (character && character.id && characterStillExists(character.id)) {
            fillSlot(slot, character);
        }
    }

    if (noiseInput) {
        noiseInput.value = restoredState.noise;
        noiseInput.addEventListener("input", saveState);
    }

    addSlotButton?.addEventListener("click", () => {
        const slots = getSlots();
        if (slots.length >= MAX_SLOTS) {
            return;
        }

        desk.appendChild(createSlot(slots.length));
        updateDeskLayout();
        saveState();
    });

    removeSlotButton?.addEventListener("click", () => {
        const slots = getSlots();
        if (slots.length <= MIN_SLOTS) {
            return;
        }

        const slot = slots[slots.length - 1];
        if (activeSlot === slot) {
            picker.close();
            activeSlot = null;
        }
        slot.remove();
        updateSlotIndexes();
        updateDeskLayout();
        saveState();
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

            saveState();
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

    document.addEventListener("click", (event) => {
        if (noiseHelp?.open && !noiseHelp.contains(event.target)) {
            noiseHelp.open = false;
        }
    });

    updateDeskLayout();

    function createSlot(index) {
        const slot = document.createElement("div");
        slot.className = "gm-slot";
        slot.dataset.slot = String(index);
        slot.innerHTML = `
            <button class="gm-slot-empty" type="button">+ Выбрать персонажа</button>
            <div class="gm-slot-filled" hidden>
                <header class="gm-slot-header">
                    <div class="gm-slot-title"></div>
                    <button class="gm-slot-remove" type="button">Убрать</button>
                </header>
                <iframe class="gm-slot-frame" title="Персонаж ${index + 1}"></iframe>
            </div>
        `;

        const emptyButton = slot.querySelector(".gm-slot-empty");
        const removeCharacterButton = slot.querySelector(".gm-slot-remove");

        emptyButton?.addEventListener("click", () => {
            activeSlot = slot;
            if (filter) {
                filter.value = "";
            }
            applyFilter("");
            picker.showModal();
            setTimeout(() => filter?.focus(), 0);
        });

        removeCharacterButton?.addEventListener("click", () => {
            clearSlot(slot);
            saveState();
        });

        return slot;
    }

    function updateSlotIndexes() {
        getSlots().forEach((slot, index) => {
            slot.dataset.slot = String(index);
            const frame = slot.querySelector(".gm-slot-frame");
            if (frame) {
                frame.title = `Персонаж ${index + 1}`;
            }
        });
    }

    function updateDeskLayout() {
        const slotCount = getSlots().length;
        desk.style.setProperty("--gm-slot-count", String(slotCount));

        if (addSlotButton) {
            addSlotButton.disabled = slotCount >= MAX_SLOTS;
        }
        if (removeSlotButton) {
            removeSlotButton.disabled = slotCount <= MIN_SLOTS;
        }
    }

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

    function saveState() {
        const slots = getSlots().map((slot) => {
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

        const state = {
            version: 2,
            slotCount: slots.length,
            noise: noiseInput?.value ?? "0",
            slots,
        };

        localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    }

    function readState() {
        let stored;
        try {
            stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || "null");
        } catch {
            stored = null;
        }

        // Совместимость со старой версией: раньше в localStorage лежал только массив из 5 слотов.
        if (Array.isArray(stored)) {
            const lastUsedIndex = stored.reduce((lastIndex, character, index) => {
                return character && character.id ? index : lastIndex;
            }, -1);

            return {
                slotCount: clamp(Math.max(DEFAULT_SLOTS, lastUsedIndex + 1), MIN_SLOTS, MAX_SLOTS),
                noise: "0",
                slots: stored.slice(0, MAX_SLOTS),
            };
        }

        if (stored && typeof stored === "object") {
            const slots = Array.isArray(stored.slots) ? stored.slots.slice(0, MAX_SLOTS) : [];
            const requestedCount = Number.parseInt(stored.slotCount, 10);
            const slotCount = Number.isFinite(requestedCount) ? requestedCount : DEFAULT_SLOTS;

            return {
                slotCount: clamp(slotCount, MIN_SLOTS, MAX_SLOTS),
                noise: stored.noise === undefined || stored.noise === null ? "0" : String(stored.noise),
                slots,
            };
        }

        return {
            slotCount: DEFAULT_SLOTS,
            noise: "0",
            slots: [],
        };
    }

    function characterStillExists(characterId) {
        return items.some((item) => item.dataset.characterId === String(characterId));
    }

    function getSlots() {
        return Array.from(desk.querySelectorAll(".gm-slot"));
    }

    function clamp(value, min, max) {
        return Math.min(max, Math.max(min, Number(value) || min));
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
