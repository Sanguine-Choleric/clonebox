const navLinks = document.querySelectorAll("nav a");
for (let i = 0; i < navLinks.length; i++) {
    var link = navLinks[i]
    if (link.getAttribute('href') === window.location.pathname) {
        link.classList.add("live");
        break;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const table = document.getElementById('split_table');

    // For type validation
    const validators = {
        iname: (value) => {
            return value.trim() !== '';
        },
        price: (value) => {
            // Number constructor checks type
            const num = Number(value);
            return !isNaN(num) && isFinite(num) && num >= 0;
        },
        quantity: (value) => {
            const num = parseInt(value, 10);
            // Check if it's an integer and not something like "12.5"
            return !isNaN(num) && isFinite(num) && num >= 0 && String(num) === value.trim();
        }
    };

    table.addEventListener('focusin', (event) => {
        if (event.target.isContentEditable) {
            event.target.dataset.originalValue = event.target.textContent.trim();
        }

    });

    table.addEventListener('keydown', (event) => {
        if (!event.target.isContentEditable) {
            return;
        }

        // Jank for preventing multiline edits
        // If enter is pressed, 'save' input
        if (event.key === 'Enter') {
            event.preventDefault();
            event.target.blur();
        }
        if (event.key === 'Escape') {
            event.target.textContent = event.target.dataset.originalValue
            event.target.blur();
        }

    });

    // Handle validation and saving on blur
    table.addEventListener('blur', (event) => {
        const cell = event.target;
        if (!cell.isContentEditable) {
            return;
        }

        const type = cell.dataset.type;
        const originalValue = cell.dataset.originalValue;
        const newValue = cell.textContent.trim();

        // Find the correct validator, if it exists
        const validator = validators[type];

        if (validator && !validator(newValue)) {
            // console.error(`Invalid input for type '${type}':`, newValue);
            cell.textContent = originalValue;
        } else {
            // if (newValue !== originalValue) {
            //     console.log(`Value changed for '${type}'. New value: ${newValue}. Sending to server...`);
            // }
        }

    }, true);

})