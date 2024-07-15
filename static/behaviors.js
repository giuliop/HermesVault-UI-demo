const behaviors = {
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
    restoreTrim: function (elem) {
        elem.value = elem.dataset.fullValue || elem.value;
    },

    /**
     * Restore the full value of all input elements in the form from the
     * data-full-value attribute (if it exists)
     */
    restoreAllTrims: function (form) {
        for (const child of form.querySelectorAll('input')) {
            child.value = child.dataset.fullValue || child.value;
        }
    },

    /**
     * trim all input elements of the form and focus on elemId
     */
    trimAllAndFocusOn: function (form, elemName) {
        for (const elem of form.querySelectorAll('input')) {
            this.trim(elem);
        }
        const targetElem = form.querySelector(`input[name="${elemName}"]`);
        if (targetElem) {
            targetElem.focus();
        }
    },
}

window.behaviors = behaviors;