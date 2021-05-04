const URL_PARAMS = new URLSearchParams(window.location.search);
let BLUR_R18 = false;

function blurElement(element, isR18) {
    if (BLUR_R18 && isR18) {
        element.classList.add("blurred")
    }
}

function makeCollapsible(collapseButton, collapsibleDiv) {
    collapseButton.addEventListener("click", function () {
        collapseButton.classList.toggle("collapse-button-active");
        if (collapsibleDiv.style.display === "block") {
            collapsibleDiv.style.display = "none";
        } else {
            collapsibleDiv.style.display = "block";
        }
    });
}

function makeAutocomplete(inputBox) {
    let currentFocus;

    function addActive(elements) {
        if (!elements) return false;

        removeActive(elements);

        if (currentFocus >= elements.length) {
            currentFocus = 0;
        }

        if (currentFocus < 0) {
            currentFocus = (elements.length - 1);
        }

        elements[currentFocus].classList.add("autocomplete-active");
    }

    function removeActive(elements) {
        for (let element of elements) {
            element.classList.remove("autocomplete-active");
        }
    }

    function closeAllLists(excludedElement) {
        let elements = document.getElementsByClassName("autocomplete-items");

        for (let i = 0; i < elements.length; i++) {
            if (elements[i] !== excludedElement && elements[i] !== inputBox) {
                elements[i].parentNode.removeChild(elements[i]);
            }
        }
    }

    let idleTimer = null;
    inputBox.addEventListener("input", function (e) {
        if (e.data === " ") return;
        clearTimeout(idleTimer);
        if (e.data == null) return;

        idleTimer = setTimeout(async function () {
            closeAllLists();
            let input = inputBox.value?.trim();
            if (!input) return false;
            currentFocus = -1;

            let listDiv = document.createElement("DIV");
            listDiv.setAttribute("id", inputBox.id + "autocomplete-list");
            listDiv.setAttribute("class", "autocomplete-items");
            inputBox.parentNode.appendChild(listDiv);

            let lastInput = input.includes(' ') ? input.split(' ').pop() : input;
            for (let match of await getMatchesFromPixiv(lastInput)) {
                if (match.name !== null && match.name !== undefined) {
                    let elementDiv = document.createElement("DIV");
                    elementDiv.innerHTML = match.name;
                    if (match.hint !== null && match.hint !== undefined) {
                        elementDiv.innerHTML += `<span style="color: grey"> ${match.hint}</span>`;
                    }
                    elementDiv.innerHTML += `<input type='hidden' value='${match.name}'>`;

                    elementDiv.addEventListener("click", function () {
                        let tag = elementDiv.getElementsByTagName("input")[0].value
                        inputBox.value = inputBox.value.replace(lastInput, tag)
                        closeAllLists();
                    });

                    listDiv.appendChild(elementDiv);
                }
            }
        }, 500)
    });

    inputBox.addEventListener("keydown", function (e) {
        let elements = document.getElementById(inputBox.id + "autocomplete-list");
        if (elements) {
            elements = elements.getElementsByTagName("div");
        }

        if (e.keyCode === 40) {
            currentFocus++;
            addActive(elements);
        } else if (e.keyCode === 38) {
            currentFocus--;
            addActive(elements);
        } else if (e.keyCode === 13) {
            if (currentFocus > -1) {
                e.preventDefault();

                if (elements) {
                    elements[currentFocus].click();
                    currentFocus = -1;
                }
            }
        }
    });

    document.addEventListener("click", function (e) {
        closeAllLists(e.target);
    });
}

async function getMatchesFromPixiv(input) {
    let matches = []

    try {
        let response = await fetch(`/autocomplete?word=${input.trimEnd()}`);
        let jsonBody = await response.json();

        for (let suggestion of jsonBody["tags"]) {
            matches.push({
                name: suggestion["name"],
                hint: suggestion["translated_name"]
            });
        }
    } catch (e) {
        console.log("Failed to retrieve autocomplete suggestions from Pixiv.", e)
    }

    return matches;
}

document.addEventListener("DOMContentLoaded", function () {
    makeAutocomplete(document.getElementById("search-box"));
    makeCollapsible(document.getElementById("advanced-options-button"), document.getElementById("advanced-options"));

    for (let dateInput of document.querySelectorAll("input[type=date]")) {
        dateInput.max = new Date().toISOString().split("T")[0];
    }

    for (let form of document.getElementsByTagName("form")) {
        form.addEventListener("submit", function () {
            for (let input of form.getElementsByTagName("input")) {
                if (input.value === "" || input.value === " " || input.value === undefined) {
                    input.setAttribute("disabled", true);
                }
            }
        });
    }

    let word = URL_PARAMS.get("word");
    if (word !== "" && word !== " " && word !== undefined) {
        document.getElementById("search-box").value = word;
    }

    let start = URL_PARAMS.get("start_date");
    if (start !== "" && start !== " " && start !== undefined) {
        document.getElementById("start_date").value = start;
    }

    let end = URL_PARAMS.get("end_date");
    if (end !== "" && end !== " " && end !== undefined) {
        document.getElementById("end_date").value = end;
    }

    let num = URL_PARAMS.get("num");
    if (num !== "" && num !== " " && num !== undefined) {
        document.getElementById("num").value = num;
    }

    if (URL_PARAMS.get("blur_r18") === "true") {
        document.getElementById("blur_r18").checked = BLUR_R18 = true;
    }

    switch (URL_PARAMS.get("search_target")) {
        case "title_and_caption":
            document.getElementById("title_and_caption").checked = true;
            break
        case "partial_match_for_tags":
            document.getElementById("partial_match_for_tags").checked = true;
            break
        default:
            document.getElementById("exact_match_for_tags").checked = true;
            break
    }

    switch (URL_PARAMS.get("sort")) {
        case "date_asc":
            document.getElementById("date_asc").checked = true;
            break
        default:
            document.getElementById("date_desc").checked = true;
            break
    }

    switch (URL_PARAMS.get("resort")) {
        case "views":
            document.getElementById("views").checked = true;
            break
        default:
            document.getElementById("bookmarks").checked = true;
            break
    }

    switch (URL_PARAMS.get("duration")) {
        case "within_last_day":
            document.getElementById("within_last_day").checked = true;
            break
        case "within_last_week":
            document.getElementById("within_last_week").checked = true;
            break
        case "within_last_month":
            document.getElementById("within_last_month").checked = true;
            break
        default:
            document.getElementById("all_time").checked = true;
            break
    }
});