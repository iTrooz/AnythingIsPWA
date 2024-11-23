let manualName = false;
let manualImage = false;

let form = document.forms["form"];

// check if we should enable autocompletion or if user entered something manually
form["name"].addEventListener("input", function() {
    manualName = !!this.value.trim();
});
form["icon_url"].addEventListener("input", function() {
    manualImage = !!this.value.trim();
    reloadIcon();
});

// Reload the icon preview
function reloadIcon() {
    let icon_url = form["icon_url"].value;

    // Image preview
    if (icon_url.trim()) { // check if it has text
        icon_preview.hidden = false;
        icon_preview.src = icon_url;
    } else {
        icon_preview.hidden = true;
    }

    // Prevent form from submitting while the icon has not loaded
    icon_preview.has_loaded = false;
}

// Check when the icon has loaded
icon_preview.addEventListener("load", function() {
    icon_preview.has_loaded = true;
    // true if image will need server processing
    let needProcessing = doesImageNeedProcessing();

    form["process_image"].value = needProcessing;
    if (needProcessing) {
        process_image_hint.innerHTML = "The icon does not meet PWA standard (png with minimum size 144x144) and will be processed to ensure it meets the requirements";
    } else {
        process_image_hint.innerHTML = "";
    }
});

// Autocomplete name and icon from the website
document.getElementById("start_url").addEventListener("input", function() {
    if (manualName&&manualImage) return;

    fetch(`/getWebsiteInfos?url=${encodeURIComponent(this.value)}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => response.json())
    .then(data => {
        if (!manualName) {
            form["name"].value = data.title || "";
        }
        if (!manualImage) {
            form["icon_url"].value = data.icon_url || "";
            reloadIcon();
        }
    })
    .catch(error => {
        console.error('Error fetching infos for website:', error);
    });
});

function isURL(s) {
    try {
        new URL(s);
        return true;
    } catch (e) {
        return false;
    }
}

// check if the image needs server-side processing
function doesImageNeedProcessing() {
    // check for png extension
    let ext = form["icon_url"].value.split('.').pop();
    if (ext != "png") return true;
    
    // Check for size
    if (icon_preview.naturalWidth < 144 || icon_preview.naturalHeight < 144) return true;

    return false;
}

// Validate form, and submit it if everything is alright
async function trySendForm() {
    let form = document.forms["form"];

    if (!form["name"].value.trim()) {
        alert("Name must be filled out");
        return;
    }else if (!isURL(form["start_url"].value)) {
        alert("URL of the website is not valid");
        return;
    } else if (!isURL(form["icon_url"].value)) {
        alert("URL of the icon is not valid");
        return;
    } else if (!icon_preview.has_loaded) {
        alert("Icon image has not loaded yet");
        return;
    }

    form["short_name"].value = form["name"].value;

    form.submit();
}
