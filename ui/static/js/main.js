const navLinks = document.querySelectorAll("nav a");
for (let i = 0; i < navLinks.length; i++) {
    var link = navLinks[i]
    if (link.getAttribute('href') === window.location.pathname) {
        link.classList.add("live");
        break;
    }
}

// Bill split upload progress. Upload + OCR can take several seconds with no other feedback
const receiptForm = document.getElementById('receipt-form')
if (receiptForm) {
    const submitBtn = receiptForm.querySelector('input[type="submit"]')
    const uploadStatus = document.getElementById('upload-status')
    const setUploading = (uploading) => {
        submitBtn.disabled = uploading
        submitBtn.value = uploading ? 'Uploading...' : 'Upload'
        uploadStatus.hidden = !uploading
    }

    // Disabling the button inside the submit handler doesn't cancel the submission; just blocks double submits
    receiptForm.addEventListener('submit', () => setUploading(true))

    // Back/forward cache restores the page as it was left, i.e. stuck mid-upload
    window.addEventListener('pageshow', (event) => {
        if (event.persisted) {
            setUploading(false)
        }
    })
}

function html_to_js_table_convert(rows, tableContents) {
    for (const row of rows) {
        const rowCells = row.querySelectorAll('td, th')
        const rowContents = []

        for (const cell of rowCells) {
            const checkboxes = cell.querySelectorAll('input[type="checkbox"]')
            if (checkboxes.length > 0) {
                const checkboxData = []
                for (const check of checkboxes) {
                    checkboxData.push(check.checked)
                }
                rowContents.push(checkboxData)
            } else {
                rowContents.push(cell.textContent.trim())
            }
        }

        tableContents.push(rowContents)
    }
}

function calc_splits(tableContents, allCheckboxes, totals, peopleNames) {
    const nameIndex = 0
    const priceIndex = 1
    const quantIndex = 2


    // Big Calc
    for (let itemIdx = 1; itemIdx < tableContents.length; itemIdx++) {
        const row = tableContents[itemIdx]
        if (!row || row.length === 0) continue

        const itemName = row[nameIndex]
        const priceRaw = row[priceIndex]
        const qtyRaw = row[quantIndex]

        // Parse price and quantity
        // TODO: Is this checking necessary?
        const itemPrice = parseFloat(String(priceRaw).replace(/[^0-9.\-]/g, '')) || 0
        const quantity = Math.max(1, parseInt(qtyRaw, 10) || 1)
        if (itemPrice <= 0) continue

        const itemTotalPrice = itemPrice * quantity

        // Checkbox matrix for this item: array[personIndex] -> array[unitIndex] -> boolean
        const perPersonUnitChecks = allCheckboxes[itemIdx - 1] || []

        for (let unit = 0; unit < quantity; unit++) {
            // Find who checked this specific unit
            const takers = []
            for (let p = 0; p < perPersonUnitChecks.length; p++) {
                const checksForPerson = perPersonUnitChecks[p] || []
                if (checksForPerson[unit]) takers.push(p)
            }

            // IMPORTANT: Non-checked units should not be charged to anyone.
            // If no one checked this unit, skip allocating its cost.
            if (takers.length === 0) {
                continue
            }

            // Split this unit's cost among takers
            const share = itemPrice / takers.length
            for (const pIdx of takers) {
                totals[peopleNames[pIdx]] += share
            }
        }
    }
}

// Person cell with one checkbox per unit of the item
function make_person_cell(quantity) {
    const cell = document.createElement('td')
    cell.className = 'person_column'
    if (quantity === 1) {
        const checkbox = document.createElement('input')
        checkbox.type = 'checkbox'
        cell.appendChild(checkbox)
    } else {
        for (let i = 0; i < quantity; i++) {
            const label = document.createElement('label')
            label.textContent = String(i + 1) + "."
            cell.appendChild(label)

            const checkbox = document.createElement('input')
            checkbox.type = 'checkbox'
            cell.appendChild(checkbox)

            const br = document.createElement('br')
            cell.appendChild(br)
        }
    }
    return cell
}

function js_totals_to_html_convert(totals) {
    const main = document.getElementsByTagName('main').item(0)

    let totalsTable = document.getElementById('totals_table')
    if (!totalsTable) {
        totalsTable = document.createElement('table')
        totalsTable.id = 'totals_table'
        totalsTable.className = 'pair_table'
    } else {
        totalsTable.replaceChildren()
    }

    // Headers - just name and total
    const headers = document.createElement('tr')
    const nameHeader = document.createElement('th')
    nameHeader.textContent = 'Name'
    headers.appendChild(nameHeader)
    const totalHeader = document.createElement('th')
    totalHeader.textContent = 'Total'
    headers.appendChild(totalHeader)
    totalsTable.appendChild(headers)

    // Adding contents of totals table into new row
    for (const name of Object.keys(totals)) {
        const row = document.createElement('tr')
        const nameCell = document.createElement('td')
        nameCell.textContent = name
        row.appendChild(nameCell)
        const totalCell = document.createElement('td')
        totalCell.textContent = totals[name]
        row.appendChild(totalCell)
        totalsTable.appendChild(row)
    }

    main.appendChild(totalsTable)
}

document.addEventListener('DOMContentLoaded', () => {
    const table = document.getElementById('split_table');
    // Only the bill split page (after an upload) has the table
    if (!table) {
        return
    }
    const addColBtn = document.getElementById('add-column-btn')
    const rmColBtn = document.getElementById('remove-column-btn')
    const addRowBtn = document.getElementById('add-row-btn')
    const rmRowBtn = document.getElementById('remove-row-btn')
    const calcBtn = document.getElementById('calc-splits-btn')

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

        const validator = validators[type];

        if (validator && !validator(newValue)) {
            // console.error(`Invalid input for type '${type}':`, newValue);
            cell.textContent = originalValue;
        } else if (type === 'quantity' && newValue !== originalValue) {
            // Rebuild person checkboxes to match the new unit count (clears this row's checks)
            const quantity = parseInt(newValue, 10)
            cell.parentElement.querySelectorAll('td.person_column').forEach(personCell => {
                personCell.replaceWith(make_person_cell(quantity))
            })
        } else {
            // if (newValue !== originalValue) {
            //     console.log(`Value changed for '${type}'. New value: ${newValue}. Sending to server...`);
            // }
        }

    }, true);

    // Col add
    addColBtn.addEventListener('click', (event) => {
        const headerRow = table.querySelector('thead tr')
        if (headerRow) {
            const newHeaderCell = document.createElement('th')
            newHeaderCell.className = 'person_column'
            newHeaderCell.content = ''
            newHeaderCell.contentEditable = 'true'
            headerRow.appendChild(newHeaderCell)
        }

        const dataRows = table.querySelectorAll('tbody tr')
        dataRows.forEach(row => {
            const quantity = parseInt(row.querySelector('td[data-type="quantity"]').textContent)
            row.appendChild(make_person_cell(quantity))
        })
    })

    // Col rm
    rmColBtn.addEventListener('click', (event) => {
        const headerRow = table.querySelector('thead tr')
        // Avoid deleting original table cols
        if (headerRow.lastElementChild.textContent === "#") {
            return
        }

        if (headerRow) {
            const lastHeaderCell = headerRow.lastElementChild
            headerRow.removeChild(lastHeaderCell)
        }
        const dataRows = table.querySelectorAll('tbody tr')
        dataRows.forEach(row => {
            const lastDataCell = row.lastElementChild
            row.removeChild(lastDataCell)
        })
    })

    // Row add, for items OCR missed
    addRowBtn.addEventListener('click', (event) => {
        const row = document.createElement('tr')
        const defaults = {iname: '', price: '0', quantity: '1'}
        for (const [type, value] of Object.entries(defaults)) {
            const cell = document.createElement('td')
            cell.dataset.type = type
            cell.contentEditable = 'true'
            cell.textContent = value
            row.appendChild(cell)
        }

        const personCount = table.querySelectorAll('thead th.person_column').length
        for (let i = 0; i < personCount; i++) {
            row.appendChild(make_person_cell(1))
        }

        table.querySelector('tbody').appendChild(row)
        row.firstElementChild.focus()
    })

    // Row rm
    rmRowBtn.addEventListener('click', (event) => {
        const lastRow = table.querySelector('tbody tr:last-child')
        if (lastRow) {
            lastRow.remove()
        }
    })

    calcBtn.addEventListener('click', (event) => {
        // Converts HTML table to a 2D array.
        // Individual checkboxes are stored as an array of booleans
        const tableContents = []
        const rows = table.querySelectorAll('tr')
        html_to_js_table_convert(rows, tableContents);

        // People names from the header: columns 3..end
        const people = []
        for (let i = 3; i < tableContents[0].length; i++) {
            people.push([tableContents[0][i]]) // keeping your original shape (array with a single name)
        }
        const peopleNames = people.map(p => p[0])
        const totals = peopleNames.reduce((acc, name) => {
            acc[name] = 0;
            return acc
        }, {})

        // Collects checkbox matrices per item row (rows 1..end, columns 3..end)
        const allCheckboxes = []
        for (let j = 1; j < tableContents.length; j++) {
            const itemCheckboxes = []
            for (let i = 3; i < tableContents[j].length; i++) {
                itemCheckboxes.push(tableContents[j][i]) // array of booleans per unit for this person
            }
            allCheckboxes.push(itemCheckboxes)
        }

        // Big Calc
        calc_splits(tableContents, allCheckboxes, totals, peopleNames);

        // Rounding math for cents
        for (const name of Object.keys(totals)) {
            // totals[name] = Math.round(totals[name] * 100) / 100
            totals[name] = totals[name].toFixed(2)
        }

        // console.log('Totals by person:', totals)

        // Converting js totals table into real html table
        js_totals_to_html_convert(totals);
    })
})