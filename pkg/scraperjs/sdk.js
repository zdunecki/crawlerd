async function keyboardType(input, text, time = 500) {
    return new Promise((resolve, reject) => {
        input.select(); // you can also use input.focus()
        input.value = "";

        let current = 0;
        const l = text.length;

        const writeText = function () {
            input.value += text[current];
            if (current < l - 1) {
                current++;
                setTimeout(function () {
                    writeText()
                }, time);
            } else {
                input.setAttribute('value', input.value);
                resolve()
            }
        }
        setTimeout(function () {
            writeText()
        }, time)
    })
}
