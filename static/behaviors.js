const Trim = {
    /**
     * Trim the value of the input element to a maximum length and display a
     * trimmed version with ellipsis. Save the full value in the data-full-value
     * attribute
     */
    trim: function (elem, maxLen = 10) {
        const halfLen = Math.floor(maxLen / 2);
        const fullValue = elem.value;
        elem.dataset.fullValue = fullValue;
        if (fullValue.length > maxLen) {
            elem.value = fullValue.slice(0, halfLen)
                + '...'
                + fullValue.slice(-halfLen);
        }
    },

    /**
     * Restore the full value of the input element from the data-full-value
     * attribute (if it exists)
     */
    restore: function (elem) {
        elem.value = elem.dataset.fullValue || elem.value;
    },

    /**
     * Restore the full value of all input elements in the form from the
     * data-full-value attribute (if it exists)
     */
    restoreAll: function (event) {
        let form = event.target;
        let formData = event.detail.parameters;
        for (let input of form.elements) {
            if (input.dataset.fullValue) {
                formData[input.name] = input.dataset.fullValue;
            }
        }
    },
}

const Show = {
    fadingTooltip: function (elem, message) {
        let tooltip = document.createElement('div');

        // Set the tooltip text
        tooltip.classList.add('tooltip');
        tooltip.textContent = message;

        // Get the bounding box of the target element
        let boundingBox = elem.getBoundingClientRect();

        // Position the tooltip
        tooltip.style.position = 'absolute';
        tooltip.style.left = `${boundingBox.left + window.scrollX}px`;
        tooltip.style.top = `${boundingBox.top + window.scrollY}px`;
        tooltip.style.display = 'block';


        // Add the tooltip to the body
        document.body.appendChild(tooltip);

        // Remove the tooltip after 2 seconds
        setTimeout(() => {
            document.body.removeChild(tooltip);
        }, 1000);;
    },
}

const behaviors = {
    Trim: Trim,
    Show: Show,
};

window.behaviors = behaviors;